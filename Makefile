#!/usr/bin/make -f

SHELL := /bin/bash

DEV_EXAMPLE := helloworld
CDK_PATH := ../cdk
TERM_PATH := ../term
TERMINFO_PATH := ../terminfo

BUILD_TAGS ?=

BUILD_TAG_OPT := $(shell \
	if [ -n "${BUILD_TAGS}" ]; then \
		echo "-tags $$(echo ${BUILD_TAGS} | perl -pe 's/\s+/,/msg')"; \
	fi \
)

RUN_ENV  ?=
RUN_ARGS ?=

DLV_PORT      ?= 2345
DLV_BIN ?= $(shell which dlv)

.PHONY: all build clean cover dev examples fmt help run dlv test tidy vet

define __go_build
$(shell \
	echo -n "#\tbuilding $(1)... "; \
	( go build -v $(3) \
			-trimpath \
			-gcflags='all="-N" -l' \
			-ldflags="\
-X 'github.com/go-curses/cdk.IncludeTtyFlag=true'  \
-X 'github.com/go-curses/cdk.IncludeProfiling=true' \
-X 'github.com/go-curses/cdk.IncludeLogFile=true'   \
-X 'github.com/go-curses/cdk.IncludeLogLevel=true'  \
" \
			${BUILD_TAG_OPT} \
			-o ./$(1) $(2) 2>&1 \
	) > ./$(1).build.log \
)
endef

define __go_build_plugin
$(call __go_build,$(1).so,$(2),-buildmode=plugin)
endef

all: help

help:
	@echo "usage: make [target]"
	@echo
	@echo "qa targets:"
	@echo "  vet         - run go vet command"
	@echo "  test        - perform all available tests"
	@echo "  cover       - perform all available tests with coverage report"
	@echo
	@echo "cleanup targets:"
	@echo "  clean       - cleans package and built files"
	@echo "  clean-logs  - cleans *.log from the project"
	@echo
	@echo "go.mod helpers:"
	@echo "  local       - add go.mod local package replacements"
	@echo "  unlocal     - remove go.mod local package replacements"
	@echo
	@echo "build targets:"
	@echo "  deps        - install stringer and bitmasker tools"
	@echo "  generate    - run go generate"
	@echo "  examples    - builds all examples"
	@echo "  build       - build test for main cdk package"
	@echo "  dev         - build ${DEV_EXAMPLE} with profiling"
	@echo "  *.so        - build a pluginworld shared object file"
	@echo "  *           - build only the given example (by name)"
	@echo
	@echo "run targets:"
	@echo "  run         - run the dev build (sanely handle crashes)"
ifneq (${DLV_BIN},)
	@echo "  dlv         - use dlv to run the dev build"
endif
	@echo "  profile.cpu - run the dev build and profile CPU"
	@echo "  profile.mem - run the dev build and profile memory"

vet:
	@echo -n "# vetting cdk ..."
	@go vet && echo " done"

test: vet
	@echo "# testing cdk ..."
	@go test -v ./...
	@for tgt in encoding env log memphis; do \
		echo "# testing cdk: $$tgt ..."; \
		cd $$tgt; \
		go test -v ./...; \
		cd - > /dev/null; \
	done
	@for tgt in `ls examples`; do \
		if [ -d examples/$$tgt ]; then \
			echo "# testing cdk example: $$tgt ..."; \
			cd examples/$$tgt; \
			go test -v ./...; \
			cd - > /dev/null; \
		fi; \
	done
	@for tgt in `ls lib`; do \
		if [ -d lib/$$tgt ]; then \
			echo "# testing cdk lib: $$tgt ..."; \
			cd lib/$$tgt; \
			go test -v ./...; \
			cd - > /dev/null; \
		fi; \
	done

cover:
	@echo "# testing cdk (with coverage) ..."
	@go test -cover -coverprofile=coverage.out ./...
	@echo "# test coverage ..."
	@go tool cover -html=coverage.out

clean-logs:
	@echo "# cleaning *.log files"
	@rm -fv *.log || true
	@echo "# cleaning *.out files"
	@rm -fv *.out || true
	@echo "# cleaning pprof files"
	@rm -rfv /tmp/*.cdk.pprof || true

clean: clean-logs
	@echo "# cleaning binaries"
	@rm -fv go_build_* || true
	@rm -fv go_test_*  || true
	@rm -fv *.so       || true
	@rm -fv demoplugin || true
	@for tgt in `ls examples`; do \
		if [ -d examples/$$tgt ]; then \
			rm -fv $$tgt || true; \
		fi; \
	done

deps:
	@echo "# installing dependencies..."
	@echo "#\tinstalling stringer"
	@GO111MODULE=off go install golang.org/x/tools/cmd/stringer
	@echo "#\tinstalling bitmasker"
	@GO111MODULE=off go install github.com/go-curses/bitmasker

generate:
	@echo "# generating go sources..."
	@go generate -v ./...

build: clean
	@echo "# building cdk"
	@go build -v ./...

examples: clean demoplugin.so demoplugin appserver hellocall helloworld mainworld pluginworld

local-term:
	@if [ -d "${TERM_PATH}" ]; then \
		echo "# adding go.mod local TERM package replacements..."; \
		go mod edit -replace=github.com/go-curses/term=${TERM_PATH}; \
	fi

unlocal-term:
	@if [ -d "${TERM_PATH}" ]; then \
		echo "# removing go.mod local TERM package replacements..."; \
		go mod edit -dropreplace=github.com/go-curses/term; \
	fi

local-terminfo:
	@if [ -d "${TERMINFO_PATH}" ]; then \
		echo "# adding go.mod local TERMINFO package replacements..."; \
		go mod edit -replace=github.com/go-curses/terminfo=${TERMINFO_PATH}; \
	fi

unlocal-terminfo:
	@if [ -d "${TERMINFO_PATH}" ]; then \
		echo "# removing go.mod local TERMINFO package replacements..."; \
		go mod edit -dropreplace=github.com/go-curses/terminfo; \
	fi

tidy:
	@echo "# running go mod tidy"
	@go mod tidy

local: local-term local-terminfo
	@echo "# adding go.mod local package replacements..."
	@go mod edit -replace=github.com/go-curses/cdk=${CDK_PATH}
	@for tgt in charset encoding env log memphis; do \
		if [ -f ${CDK_PATH}/$$tgt/go.mod ]; then \
			echo "#\t$$tgt"; \
			go mod edit -replace=github.com/go-curses/cdk/$$tgt=${CDK_PATH}/$$tgt ; \
		fi; \
	done
	@for tgt in `ls ${CDK_PATH}/lib`; do \
		if [ -f ${CDK_PATH}/lib/$$tgt/go.mod ]; then \
			echo "#\tlib/$$tgt"; \
			go mod edit -replace=github.com/go-curses/cdk/lib/$$tgt=${CDK_PATH}/lib/$$tgt ; \
		fi; \
	done

unlocal: unlocal-term unlocal-terminfo
	@echo "# removing go.mod local CDK package replacements..."
	@go mod edit -dropreplace=github.com/go-curses/cdk
	@for tgt in charset encoding env log memphis; do \
		if [ -f ${CDK_PATH}/$$tgt/go.mod ]; then \
			echo "#\t$$tgt"; \
			go mod edit -dropreplace=github.com/go-curses/cdk/$$tgt ; \
		fi; \
	done
	@for tgt in `ls ${CDK_PATH}/lib`; do \
		if [ -f ${CDK_PATH}/lib/$$tgt/go.mod ]; then \
			echo "#\tlib/$$tgt"; \
			go mod edit -dropreplace=github.com/go-curses/cdk/lib/$$tgt ; \
		fi; \
	done

dev: clean
	@if [ -d examples/${DEV_EXAMPLE} ]; \
	then \
		echo -n "# building: ${DEV_EXAMPLE} [dev]... "; \
		$(call __go_build,${DEV_EXAMPLE},./examples/${DEV_EXAMPLE}); \
		[ -f ${DEV_EXAMPLE} ] \
			&& echo "done." \
			|| echo -e "failed.\n>\tsee ./${DEV_EXAMPLE}.build.log for errors"; \
	else \
		echo "# dev example not found: ${DEV_EXAMPLE}"; \
	fi

run: export GO_CDK_LOG_FILE=./${DEV_EXAMPLE}.cdk.log
run: export GO_CDK_LOG_LEVEL=debug
run: export GO_CDK_LOG_FULL_PATHS=true
run:
	@if [ -f ${DEV_EXAMPLE} ]; \
	then \
		echo "# starting ${DEV_EXAMPLE}..."; \
		( ${RUN_ENV} ./${DEV_EXAMPLE} ${RUN_ARGS} 2>> "${GO_CDK_LOG_FILE}" ); \
		if [ $$? -ne 0 ]; \
		then \
			stty sane; echo ""; \
			echo "# ${DEV_EXAMPLE} crashed, see: ./${DEV_EXAMPLE}.cdk.log"; \
			read -p "# reset terminal? [Yn] " RESP; \
			if [ "$$RESP" = "" -o "$$RESP" = "Y" -o "$$RESP" = "y" ]; \
			then \
				reset; \
				echo "# ${DEV_EXAMPLE} crashed, terminal reset, see: ./${DEV_EXAMPLE}.cdk.log"; \
			fi; \
		else \
			echo "# ${DEV_EXAMPLE} exited normally."; \
		fi; \
	fi

ifneq (${DLV_BIN},)
dlv: export GO_CDK_LOG_FILE=./${DEV_EXAMPLE}.cdk.log
dlv: export GO_CDK_LOG_LEVEL=debug
dlv: export GO_CDK_LOG_FULL_PATHS=true
dlv:
	@if [ -f ${DEV_EXAMPLE} ]; \
	then \
		echo "# delving ${DEV_EXAMPLE}..."; \
		( ${RUN_ENV} \
			${DLV_BIN} --listen=:${DLV_PORT} --headless=true \
			--api-version=2 --accept-multiclient exec -- \
			./${DEV_EXAMPLE} ${RUN_ARGS} \
			2>> "${GO_CDK_LOG_FILE}" \
		); \
		if [ $$? -ne 0 ]; \
		then \
			stty sane; echo ""; \
			echo "# ${DEV_EXAMPLE} crashed, see: ./${DEV_EXAMPLE}.cdk.log"; \
			read -p "# reset terminal? [Yn] " RESP; \
			if [ "$$RESP" = "" -o "$$RESP" = "Y" -o "$$RESP" = "y" ]; \
			then \
				reset; \
				echo "# ${DEV_EXAMPLE} crashed, terminal reset, see: ./${DEV_EXAMPLE}.cdk.log"; \
			fi; \
		else \
			echo "# ${DEV_EXAMPLE} exited normally."; \
		fi; \
	fi
endif

profile.cpu: export GO_CDK_LOG_FILE=./${DEV_EXAMPLE}.cdk.log
profile.cpu: export GO_CDK_LOG_LEVEL=debug
profile.cpu: export GO_CDK_LOG_FULL_PATHS=true
profile.cpu: export GO_CDK_PROFILE_PATH=/tmp/${DEV_EXAMPLE}.cdk.pprof
profile.cpu: export GO_CDK_PROFILE=cpu
profile.cpu: dev
	@mkdir -v /tmp/${DEV_EXAMPLE}.cdk.pprof 2>/dev/null || true
	@if [ -f ${DEV_EXAMPLE} ]; \
		then \
			( ${RUN_ENV} ./${DEV_EXAMPLE} ${RUN_ARGS} 2>> "${GO_CDK_LOG_FILE}" ); \
			if [ $$? -eq 0 ]; then \
				if [ -f /tmp/${DEV_EXAMPLE}.cdk.pprof/cpu.pprof ]; \
				then \
					read -p "# Press enter to open a pprof instance" JUNK \
					&& go tool pprof -http=:8080 /tmp/${DEV_EXAMPLE}.cdk.pprof/cpu.pprof ; \
				else \
					echo "# missing /tmp/${DEV_EXAMPLE}.cdk.pprof/cpu.pprof"; \
				fi ; \
			fi ; \
		fi

profile.mem: export GO_CDK_LOG_FILE=./${DEV_EXAMPLE}.log
profile.mem: export GO_CDK_LOG_LEVEL=debug
profile.mem: export GO_CDK_LOG_FULL_PATHS=true
profile.mem: export GO_CDK_PROFILE_PATH=/tmp/${DEV_EXAMPLE}.cdk.pprof
profile.mem: export GO_CDK_PROFILE=mem
profile.mem: dev
	@mkdir -v /tmp/${DEV_EXAMPLE}.cdk.pprof 2>/dev/null || true
	@if [ -f ${DEV_EXAMPLE} ]; \
		then \
			( ${RUN_ENV} ./${DEV_EXAMPLE} ${RUN_ARGS} 2>> "${GO_CDK_LOG_FILE}" ); \
			if [ $$? -eq 0 ]; then \
				if [ -f /tmp/${DEV_EXAMPLE}.cdk.pprof/mem.pprof ]; \
				then \
					read -p "# Press enter to open a pprof instance" JUNK \
					&& go tool pprof -http=:8080 /tmp/${DEV_EXAMPLE}.cdk.pprof/mem.pprof; \
				else \
					echo "# missing /tmp/${DEV_EXAMPLE}.cdk.pprof/mem.pprof"; \
				fi ; \
			fi ; \
		fi

%.so: PLUGNAME=$(basename $@)
%.so:
	@if [ -d "examples/plugin-world/$(PLUGNAME)" ]; \
	then \
		echo -n "# building plugin $(PLUGNAME)... "; \
		$(call __go_build_plugin,$(PLUGNAME),./examples/pluginworld/$(PLUGNAME)); \
		[ -f $@ ] \
			&& echo "done." \
			|| echo -e "fail.\n#\tsee ./$(PLUGNAME).build.log"; \
	else \
		echo "not a plugin: $@"; \
		false; \
	fi

%:
	@if [ -d "examples/$@" ]; then \
		echo -n "# building example $@... "; \
		$(call __go_build,$@,./examples/$@); \
		[ -f $@ ] \
			&& echo "done." \
			|| echo -e "fail.\n#\tsee ./$@.build.log"; \
	elif [ -d "examples/pluginworld/$@"]; \
	then \
		echo -n "# building example pluginworld/$@... "; \
		$(call __go_build,$@,./examples/pluginworld/$@); \
		[ -f $@ ] \
			&& echo "done." \
			|| echo -e "fail.\n#\tsee ./$@.build.log"; \
	else \
		echo "not an example: $@"; \
	fi

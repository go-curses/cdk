#!/usr/bin/make -f

SHELL := /bin/bash

DEV_EXAMPLE := helloworld
CDK_PATH := ../cdk
TERM_PATH := ../term
TERMINFO_PATH := ../terminfo

.PHONY: all build clean cover dev examples fmt help run test tidy vet

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
	@go build -v

examples: clean demoplugin.so demoplugin
	@echo "# building all examples..."
	@for name in `ls examples`; do \
		if [ -d examples/$$name ]; then \
			cd examples/$$name/; \
			echo -n "#\tbuilding $$name... "; \
			( go build -v \
					-trimpath \
					-gcflags=all="-N -l" \
					-ldflags="\
-X 'main.IncludeProfiling=true' \
-X 'main.IncludeLogFile=true'   \
-X 'main.IncludeLogLevel=true'  \
" \
					-o ../../$$name 2>&1 \
			) > ../../$$name.build.log; \
			cd - > /dev/null; \
			[ -f $$name ] \
				&& echo "done." \
				|| echo "failed.\n>\tsee ./$$name.build.log for errors"; \
		fi; \
	done

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
		cd examples/${DEV_EXAMPLE}; \
		( go build -v \
				-ldflags="\
-X 'main.IncludeProfiling=true' \
-X 'main.IncludeLogFile=true'   \
-X 'main.IncludeLogLevel=true'  \
" \
				-o ../../${DEV_EXAMPLE} \
			2>&1 ) > ../../${DEV_EXAMPLE}.build.log; \
		cd - > /dev/null; \
		[ -f ${DEV_EXAMPLE} ] \
			&& echo "done." \
			|| echo "failed.\n>\tsee ./${DEV_EXAMPLE}.build.log for errors"; \
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
		./${DEV_EXAMPLE}; \
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

profile.cpu: export GO_CDK_LOG_FILE=./${DEV_EXAMPLE}.cdk.log
profile.cpu: export GO_CDK_LOG_LEVEL=debug
profile.cpu: export GO_CDK_LOG_FULL_PATHS=true
profile.cpu: export GO_CDK_PROFILE_PATH=/tmp/${DEV_EXAMPLE}.cdk.pprof
profile.cpu: export GO_CDK_PROFILE=cpu
profile.cpu: dev
	@mkdir -v /tmp/${DEV_EXAMPLE}.cdk.pprof 2>/dev/null || true
	@if [ -f ${DEV_EXAMPLE} ]; \
		then \
			./${DEV_EXAMPLE} && \
			if [ -f /tmp/${DEV_EXAMPLE}.cdk.pprof/cpu.pprof ]; \
			then \
				read -p "# Press enter to open a pprof instance" JUNK \
				&& go tool pprof -http=:8080 /tmp/${DEV_EXAMPLE}.cdk.pprof/cpu.pprof ; \
			else \
				echo "# missing /tmp/${DEV_EXAMPLE}.cdk.pprof/cpu.pprof"; \
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
			./${DEV_EXAMPLE} && \
			if [ -f /tmp/${DEV_EXAMPLE}.cdk.pprof/mem.pprof ]; \
			then \
				read -p "# Press enter to open a pprof instance" JUNK \
				&& go tool pprof -http=:8080 /tmp/${DEV_EXAMPLE}.cdk.pprof/mem.pprof; \
			else \
				echo "# missing /tmp/${DEV_EXAMPLE}.cdk.pprof/mem.pprof"; \
			fi ; \
		fi

%.so: PLUGNAME=$(basename $@)
%.so:
	@if [ -d examples/pluginworld/$(PLUGNAME) ]; \
	then \
		echo -n "# building plugin $(PLUGNAME)... "; \
		cd examples/pluginworld/$(PLUGNAME); \
		( go build -v \
				-buildmode=plugin \
				-trimpath \
				-gcflags=all="-N -l" \
				-ldflags="\
-X 'main.IncludeProfiling=true' \
-X 'main.IncludeLogFile=true'   \
-X 'main.IncludeLogLevel=true'  \
" \
				-o ../../../$@ \
			2>&1 ) > ../../../$(PLUGNAME).build.log; \
		cd - > /dev/null; \
		if [ -f $@ ]; \
		then \
			echo "done."; \
		else \
			echo "fail.\n#\tsee ./$(PLUGNAME).build.log"; \
		fi; \
	else \
		echo "not a plugin: $@"; \
		false; \
	fi

%:
	@if [ -d examples/$@ ]; \
	then \
		echo -n "# building example $@... "; \
		cd examples/$@; \
		( go build -v \
				-trimpath \
				-gcflags=all="-N -l" \
				-ldflags="\
-X 'main.IncludeProfiling=true' \
-X 'main.IncludeLogFile=true'   \
-X 'main.IncludeLogLevel=true'  \
" \
				-o ../../$@ \
			2>&1 ) > ../../$@.build.log; \
		cd - > /dev/null; \
		if [ -f $@ ]; \
		then \
			echo "done."; \
		else \
			echo "fail.\n#\tsee ./$@.build.log"; \
		fi; \
	elif [ -d examples/pluginworld/$@ ]; \
	then \
		echo -n "# building example pluginworld/$@... "; \
		cd examples/pluginworld/$@; \
		( go build -v \
				-trimpath \
				-gcflags=all="-N -l" \
				-ldflags="\
-X 'main.IncludeProfiling=true' \
-X 'main.IncludeLogFile=true'   \
-X 'main.IncludeLogLevel=true'  \
" \
				-o ../../../$@ \
			2>&1 ) > ../../../$@.build.log; \
		cd - > /dev/null; \
		if [ -f $@ ]; \
		then \
			echo "done."; \
		else \
			echo "fail.\n#\tsee ./$@.build.log"; \
		fi; \
	else \
		echo "not an example: $@"; \
	fi

package bits

/*
widget State Types
*/

type BitFlags interface {
	Get() CBitFlags
	Set(f CBitFlags) BitFlags
	Clear(f CBitFlags) BitFlags
	Toggle(f CBitFlags) BitFlags
	Has(f CBitFlags) bool
}

type CBitFlags uint64

func (b CBitFlags) Set(f CBitFlags) CBitFlags {
	return b | f
}

func (b CBitFlags) Clear(f CBitFlags) CBitFlags {
	return b &^ f
}

func (b CBitFlags) Toggle(f CBitFlags) CBitFlags {
	return b ^ f
}

func (b CBitFlags) Has(f CBitFlags) bool {
	return b&f != 0
}

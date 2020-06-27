package factory

type extendFieldInfo struct {
	name string

	typeName string

	typePath string
}

func newLoggerFieldInfo() extendFieldInfo {
	efi := extendFieldInfo{
		name:     "logger",
		typeName: "Logger",
		typePath: "github.com/jukylin/esim/log",
	}

	return efi
}

func newConfigFieldInfo() extendFieldInfo {
	efi := extendFieldInfo{
		name:     "conf",
		typeName: "Config",
		typePath: "github.com/jukylin/esim/config",
	}

	return efi
}

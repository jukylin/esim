package file_dir


type IfaceWriter interface {
	Write(outFile, content string) error
}

type NullWrite struct{}

func NewNullWrite() IfaceWriter {
	return &NullWrite{}
}

func (nw *NullWrite) Write(outFile, content string) error { return nil }


type EsimWriter struct{}

func NewEsimWriter() IfaceWriter {
	return &EsimWriter{}
}

func (ew *EsimWriter) Write(outFile, content string) error {
	return EsimWrite(outFile, content)
}

package filedir

import "errors"

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

// ErrWrite for write errors.
type ErrWrite struct {
	nilNum int

	count int
}

func NewErrWrite(nilNum int) IfaceWriter {
	return &ErrWrite{nilNum: nilNum}
}

func (er *ErrWrite) Write(outFile, content string) error {
	er.count++
	if er.nilNum > 0 && er.nilNum >= er.count {
		return nil
	}

	return errors.New("write error")
}

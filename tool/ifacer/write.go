package ifacer

import "github.com/jukylin/esim/pkg/file-dir"

type IfaceWrite interface {
	Write(outFile, content string) error
}

type NullWrite struct{}

func (this NullWrite) Write(outFile, content string) error { return nil }

type EsimWrite struct{}

func (this EsimWrite) Write(outFile, content string) error {
	return file_dir.EsimWrite(outFile, content)
}

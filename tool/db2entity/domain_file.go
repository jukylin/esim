package db2entity

import "github.com/spf13/viper"

//
//type domainFile struct {
//	tmplName string
//
//	data interface{}
//
//	tpl string
//
//	writePath string
//}

type domainFile interface {
	cloumnsToTplData(columns []columns) interface{}
}


type entityDomainFile struct {
	writeTarget string

	tplName string

	template string

	data string
}

func NewEntityDomainFile(v *viper.Viper) *entityDomainFile {

	edf := &entityDomainFile{}

	edf.bindInput(v)

	return edf
}

func (edf *entityDomainFile) bindInput(v *viper.Viper)  {

}


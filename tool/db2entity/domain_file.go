package db2entity

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

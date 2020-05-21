package new

func init() {
	Files = append(Files, mainfc)
}

var (
	mainfc = &FileContent{
		FileName: "main.go",
		Dir:      ".",
		Content: `package main

import (
	"{{.ProPath}}{{.ServerName}}/cmd"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func main() {
  cmd.Execute()
}
`,
	}
)

package new

func init()  {
	Files = append(Files, mainfc)
}

var (
	mainfc = &FileContent{
		FileName: "main.go",
		Dir:      ".",
		Content: `package main

import "{{PROPATH}}{{service_name}}/cmd"

func main() {
  cmd.Execute()
}
`,
	}

)

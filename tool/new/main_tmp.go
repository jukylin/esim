package new

func MainInit() {
	fc := &FileContent{
		FileName: "main.go",
		Dir:      ".",
		Content: `package main

import "{{PROPATH}}{{service_name}}/cmd"

func main() {
  cmd.Execute()
}
`,
	}

	Files = append(Files, fc)
}

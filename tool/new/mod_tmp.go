package new

func ModInit() {
	fc := &FileContent{
		FileName: "go.mod",
		Dir:      ".",
		Content: `module {{PROPATH}}{{service_name}}
`,
	}

	Files = append(Files, fc)
}

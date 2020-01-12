package new

func GitIgnoreInit() {
	fc := &FileContent{
		FileName: ".gitignore",
		Dir:      ".",
		Content: `/{{service_name}}
lastupdate.tmp
*.tar.gz
.com.apple*
.idea
*.svg
*.proto
`,
	}

	Files = append(Files, fc)
}

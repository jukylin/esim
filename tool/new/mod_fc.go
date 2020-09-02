package new

var (
	modfc = &FileContent{
		FileName: "go.mod",
		Dir:      ".",
		Content:  `module {{.ProPath}}{{.ServerName}}`,
	}
)

func initModFiles() {
	Files = append(Files, modfc)
}

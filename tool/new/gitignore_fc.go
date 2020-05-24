package new

var (
	gitIgnorefc = &FileContent{
		FileName: ".gitignore",
		Dir:      ".",
		Content: `/{{.ServerName}}
lastupdate.tmp
*.tar.gz
.com.apple*
.idea
*.svg
*.proto`,
	}
)

func initGitIgnoreFiles() {
	Files = append(Files, gitIgnorefc)
}

package new

var (
	domainfc1 = &FileContent{
		FileName: "user.go",
		Dir:      "internal/domain/user/entity",
		Content: `package entity

type User struct {

	// ID
	ID int {{.SingleMark}}gorm:"column:id;primary_key"{{.SingleMark}}

	// username
	UserName string {{.SingleMark}}gorm:"column:user_name"{{.SingleMark}}

	// pwd
	PassWord string {{.SingleMark}}gorm:"column:pass_word"{{.SingleMark}}
}
`,
	}

	domainfc2 = &FileContent{
		FileName: "README.md",
		Dir:      "internal/domain/user/service",
		Content:  `domain service`,
	}
)

func initDomainFiles() {
	Files = append(Files, domainfc1, domainfc2)
}

package new

func init()  {
	Files = append(Files, domainfc1, domainfc2)
}

var(
	domainfc1 = &FileContent{
		FileName: "user_entity.go",
		Dir:      "internal/domain/user/entity",
		Content: `package entity

type User struct {

	//ID
	ID int {{.SingleMark}}gorm:"column:id;primary_key"{{.SingleMark}}

	//优惠券号码
	UserName string {{.SingleMark}}gorm:"column:user_name"{{.SingleMark}}

	//密码
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

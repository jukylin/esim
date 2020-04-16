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
	ID int {{!}}gorm:"column:id;primary_key"{{!}}

	//优惠券号码
	UserName string {{!}}gorm:"column:user_name"{{!}}

	//密码
	PassWord string {{!}}gorm:"column:pass_word"{{!}}
}
`,
	}

	domainfc2 = &FileContent{
		FileName: "README.md",
		Dir:      "internal/domain/user/service",
		Content:  `domain service`,
	}

)

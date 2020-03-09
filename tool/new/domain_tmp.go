package new

func ModelInit() {
	fc1 := &FileContent{
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

	fc2 := &FileContent{
		FileName: "README.md",
		Dir:      "internal/domain/user/service",
		Content:  `domain service`,
	}

	Files = append(Files, fc1, fc2)
}

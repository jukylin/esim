package new

func ModelInit() {
	fc1 := &FileContent{
		FileName: "user.go",
		Dir:      "internal/domain/entity",
		Content: `package entity

type User struct {

	//ID
	ID int {{!}}gorm:"column:id;primary_key"{{!}}

	//优惠券号码
	UserNmae string {{!}}gorm:"column:user_name"{{!}}

	//密码
	PassWord string {{!}}gorm:"column:pass_word"{{!}}
}
`,
	}

	fc2 := &FileContent{
		FileName: "README.md",
		Dir:      "internal/domain/dto",
		Content:  `for DTO`,
	}

	fc3 := &FileContent{
		FileName: "user.go",
		Dir:      "internal/domain/dto",
		Content: `package dto

import "{{PROPATH}}{{service_name}}/internal/domain/entity"

type User struct {

	//用户名称
	UserName string {{!}}json:"user_name"{{!}}

	//密码
	PassWord string {{!}}json:"pass_word"{{!}}
}

func NewUser(user entity.User) User {
	dto := User{}

	dto.UserName = user.UserNmae
	dto.PassWord = user.PassWord
	return dto
}`,
	}

	Files = append(Files, fc1, fc2, fc3)
}

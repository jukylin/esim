package new

func init()  {
	Files = append(Files, thirdPartyfc1, thirdPartyfc2)
}

var (
	thirdPartyfc1 = &FileContent{
		FileName: "README.md",
		Dir:      "internal/infra/third_party",
		Content:  `用于存放第三方文件`,
	}

	thirdPartyfc2 = &FileContent{
		FileName: "README.md",
		Dir:      "internal/infra/third_party/protobuf",
		Content:  `用于存放 proto 生成的源码文件`,
	}

)

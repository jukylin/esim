package new

func ThirdPartyInit() {
	fc1 := &FileContent{
		FileName: "README.md",
		Dir:      "internal/infra/third_party",
		Content:  `用于存放第三方文件`,
	}

	fc2 := &FileContent{
		FileName: "README.md",
		Dir:      "internal/infra/third_party/protobuf",
		Content:  `用于存放 proto 生成的源码文件`,
	}

	Files = append(Files, fc1, fc2)
}

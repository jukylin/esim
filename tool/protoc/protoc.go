package protoc

import (
	"bufio"
	"fmt"
	"github.com/spf13/viper"
	"github.com/jukylin/esim/pkg/file-dir"
	logger "github.com/jukylin/esim/log"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	log logger.Logger
)

func init() {
	log = logger.NewLogger()
}

func Gen(v *viper.Viper) {

	target := v.GetString("target")

	ex, err := file_dir.IsExistsDir(target)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	if ex == false {
		log.Errorf(target + " 目录不存在")
		return
	}

	from_proto := v.GetString("from_proto")
	if from_proto == "" {
		log.Errorf("没有指定proto文件")
		return
	}

	pkg_name := v.GetString("package")
	if pkg_name == "" {
		pkg_name = getPkgName(from_proto)
	}

	err = file_dir.CreateDir(target + "/" + pkg_name)
	if err != nil {
		log.Errorf("创建 "+target+"/"+pkg_name+" 失败", err.Error())
		return
	}

	log.Infof("请确认已安装protoc")
	pwd, _ := os.Getwd()

	cmd_line := fmt.Sprintf("protoc --go_out=plugins=grpc:%s %s", target+"/"+pkg_name, from_proto)

	log.Infof(cmd_line)

	args := strings.Split(cmd_line, " ")

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = pwd

	cmd.Env = os.Environ()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Errorf(err.Error())
	}

	if v.GetBool("mock") == true {
		log.Infof("请确认已安装 mockery")

		_, fileName := filepath.Split(from_proto)
		fileStrs := strings.Split(fileName, ".")
		file_name := fileStrs[0]

		destination := target + "/" + pkg_name + "/mock_" + file_name + ".go"
		source := target + "/" + pkg_name + "/" + file_name + ".pb.go"

		mockgen_cmd_line := fmt.Sprintf("mockgen -destination %s -package %s -source %s",
			destination, pkg_name, source)

		log.Infof(mockgen_cmd_line)

		args := strings.Split(mockgen_cmd_line, " ")
		cmdMock := exec.Command(args[0], args[1:]...)
		cmdMock.Dir = pwd

		cmdMock.Env = os.Environ()

		cmdMock.Stdout = os.Stdout
		cmdMock.Stderr = os.Stderr

		err = cmdMock.Run()
		if err != nil {
			log.Errorf(err.Error())
		}
	}

	return
}

func getPkgName(proto_file string) string {
	f, err := os.Open(proto_file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var pkgName string

	rd := bufio.NewReader(f)
	for {
		line, err := rd.ReadString('\n') //以'\n'为结束符读入一行

		if err != nil || io.EOF == err {
			break
		}

		match, err := regexp.MatchString("^package", line)
		if err != nil {
			log.Errorf(err.Error())
		}
		if match {
			reg, err := regexp.Compile(`\w+`)
			if err != nil {
				log.Errorf(err.Error())
			} else {
				strs := reg.FindAllString(line, -1)
				return strs[1]
			}
		}
	}

	return pkgName
}

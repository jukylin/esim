package new

import (
	"errors"
	"github.com/spf13/viper"
	"github.com/jukylin/esim/pkg/file_dir"
	logger "github.com/jukylin/esim/log"
	"io/ioutil"
	"os"
	"strings"
)

var (
	Files []*FileContent
)

type FileContent struct {
	FileName string `json:"file_name"`
	Dir      string `json:"dir"`
	Content  string `json:"context"`
}

func Build(v *viper.Viper, log logger.Logger) error {

	var err error
	serviceName := v.GetString("service_name")
	if serviceName == "" {
		return errors.New("请输入 service_name")
	}

	exists, err := file_dir.IsExistsDir("./" + serviceName)
	if exists {
		return errors.New("创建目录 " + serviceName + " 失败：目录已存在")
	}

	if err != nil {
		return errors.New("检查目录失败：" + err.Error())
	}

	err = file_dir.CreateDir(serviceName)
	if err != nil {
		return errors.New("创建目录 " + serviceName + " 失败：" + err.Error())
	}

	tmpInit(v)

	currentDir := file_dir.GetCurrentDir()
	if currentDir != ""{
		currentDir = currentDir + "/"
	}
	for _, file := range Files {
		dir := serviceName + "/" + file.Dir

		exists, err := file_dir.IsExistsDir(dir)
		if err != nil {
			//半路创建失败，全删除
			os.Remove(serviceName)
			return err
		}

		if exists == false {
			err = file_dir.CreateDir(dir)
			if err != nil {
				//半路创建失败，全删除
				os.Remove(serviceName)
				return err
			}
		}
		fileName := dir + "/" + file.FileName

		file.Content = strings.ReplaceAll(file.Content, "{{service_name}}", serviceName)
		file.Content = strings.ReplaceAll(file.Content, "{{!}}", "`")
		file.Content = strings.ReplaceAll(file.Content, "{{PROPATH}}", currentDir)

		if file.FileName == "monitoring.yaml" {
			if v.GetBool("monitoring") == false {
				log.Infof("关闭监控")
				file.Content = strings.ReplaceAll(file.Content, "{{bool}}", "false")
			} else {
				log.Infof("开启监控")
				file.Content = strings.ReplaceAll(file.Content, "{{bool}}", "true")
			}
		}

		//写内容
		err = ioutil.WriteFile(fileName, []byte(file.Content), 0666)
		if err != nil {
			//半路创建失败，全删除
			os.Remove(serviceName)
			return err
		}
		log.Infof(fileName + " 写入完成")
	}

	return nil
}

func tmpInit(v *viper.Viper) {
	CmdInit()
	ConfigInit()
	DaoInit()
	GitIgnoreInit()
	MainInit()
	ModInit()
	ModelInit()
	ProtoBufInit()
	ServiceInit()
	ThirdPartyInit()
	InfraInit()
	RepoInit()

	if v.GetBool("gin") == true {
		GinInit()
	}

	if v.GetBool("beego") == true {
		BeegoInit()
	}

	if v.GetBool("grpc") == true {
		GrpcInit()
	}
}

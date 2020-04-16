package new

import (
	"errors"
	"github.com/spf13/viper"
	"github.com/jukylin/esim/pkg/file-dir"
	logger "github.com/jukylin/esim/log"
	"io/ioutil"
	"os"
	"strings"
	"golang.org/x/tools/imports"
	"path/filepath"
	"regexp"
)

var (
	Files = make([]*FileContent, 0)
)

type FileContent struct {
	FileName string `json:"file_name"`
	Dir      string `json:"dir"`
	Content  string `json:"context"`
}

type Project struct {
	ServerName string

	PackageName string

	RunServer string

	ProPath string

	ImportServer string

	//true or false string type
	Monitoring string

	logger logger.Logger

	withGin bool

	withBeego bool

	withGrpc bool
}

func NewProject(logger logger.Logger) *Project {

	project := &Project{}

	project.logger = logger

	return project
}

func Build(v *viper.Viper) error {

	var err error
	serviceName := v.GetString("service_name")
	if serviceName == "" {
		return errors.New("请输入 service_name")
	}

	if strings.Contains(serviceName, ".") == true{
		return errors.New("服务名称不能包含【.】")
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



	// "-" => "_"
	packageName := strings.Replace(serviceName, "-", "_", -1)

	run_server, import_server := tmpInit(v)

	currentDir := file_dir.GetGoProPath()
	if currentDir != ""{
		currentDir = currentDir + string(filepath.Separator)
	}
	for _, file := range Files {
		dir := serviceName + string(filepath.Separator) + file.Dir

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
		fileName := dir + string(filepath.Separator) + file.FileName
		//before all replace
		file.Content = strings.ReplaceAll(file.Content, "{{IMPORT_SERVER}}", import_server)

		file.Content = strings.ReplaceAll(file.Content, "{{service_name}}", serviceName)

		file.Content = strings.ReplaceAll(file.Content, "{{package_name}}", packageName)

		file.Content = strings.ReplaceAll(file.Content, "{{!}}", "`")
		file.Content = strings.ReplaceAll(file.Content, "{{PROPATH}}", currentDir)

		file.Content = strings.ReplaceAll(file.Content, "{{RUN_SERVER}}", run_server)

		if file.FileName == "monitoring.yaml" {
			if v.GetBool("monitoring") == false {
				//log.Infof("关闭监控")
				file.Content = strings.ReplaceAll(file.Content, "{{bool}}", "false")
			} else {
				//log.Infof("开启监控")
				file.Content = strings.ReplaceAll(file.Content, "{{bool}}", "true")
			}
		}

		var src []byte
		if strings.HasSuffix(fileName, ".go") {
			src, err = imports.Process("", []byte(file.Content), nil)
			if err != nil {
				os.Remove(serviceName)
				return err
			}
		}else{
			src = []byte(file.Content)
		}

		//写内容
		err = ioutil.WriteFile(fileName, src, 0666)
		if err != nil {
			//半路创建失败，全删除
			os.Remove(serviceName)
			return err
		}
		//log.Infof(fileName + " 写入完成")
	}

	return nil
}

func (this *Project) Run(v *viper.Viper) {
	this.bindInput(v)

	this.createServer()

	this.getPackName()
}

func (this *Project) bindInput(v *viper.Viper) bool {
	var err error
	serverName := v.GetString("server_name")
	if serverName == "" {
		this.logger.Fatalf("The server_name is empty")
	}
	this.ServerName = serverName

	if this.checkServerName() == false {
		this.logger.Fatalf("The server_name only supports【a-z_-】")
	}

	exists, err := file_dir.IsExistsDir(serverName)
	if err != nil {
		this.logger.Fatalf(err.Error())
	}

	if exists {
		this.logger.Fatalf("The %s is exists can't be create", serverName)
	}

	err = file_dir.CreateDir(serverName)
	if err != nil {
		this.logger.Fatalf(err.Error())
	}

	this.withGin = v.GetBool("gin")

	this.withBeego = v.GetBool("beego")

	if this.withGin == true && this.withBeego == true {
		this.logger.Fatalf("either gin or beego")
	}

	this.withGrpc = v.GetBool("grpc")

	return true
}

func (this *Project) checkServerName() bool {
	match, err := regexp.MatchString("^[a-z_-]+$", this.ServerName)
	if err != nil {
		this.logger.Fatalf(err.Error())
	}

	return match
}

func (this *Project) createServer() bool {
	match, err := regexp.MatchString("^[a-z_-]+$", this.ServerName)
	if err != nil {
		this.logger.Fatalf(err.Error())
	}

	return match
}

//PackName In most cases,  ServerName eq PackageName
func (this *Project) getPackName() {
	this.PackageName = strings.Replace(this.ServerName, "-", "_", -1)
}

func tmpInit(v *viper.Viper) (string, string) {

	var runServer string
	var importServer string

	if v.GetBool("gin") == true {
		//GinInit()
		runServer = "		app.Trans = append(app.Trans, http.NewGinServer(app))\n"
		importServer += "	\"{{PROPATH}}{{service_name}}/internal/transports/http\"\n"
	}

	if v.GetBool("beego") == true {
		//BeegoInit()
		runServer += "		app.Trans = append(app.Trans, http.NewBeegoServer(app.Esim))\n"
		importServer += "	\"{{PROPATH}}{{service_name}}/internal/transports/http\"\n"
	}

	if v.GetBool("grpc") == true {
		//GrpcInit()
		runServer += "		app.Trans = append(app.Trans, grpc.NewGrpcServer(app))\n"
		importServer += "	\"{{PROPATH}}{{service_name}}/internal/transports/grpc\"\n"
	}

	return runServer, importServer
}

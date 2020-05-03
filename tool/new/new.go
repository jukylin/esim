package new

import (
	"path/filepath"
	"regexp"
	"strings"

	logger "github.com/jukylin/esim/log"
	file_dir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/pkg/templates"
	"github.com/spf13/viper"
	"golang.org/x/tools/imports"
)

var (
	Files = make([]*FileContent, 0)
)

type FileContent struct {
	FileName string `json:"file_name"`
	Dir      string `json:"dir"`
	Content  string `json:"context"`
}

//Project
type Project struct {
	ServerName string

	PackageName string

	RunTrans []string

	ProPath string

	ImportServer []string

	SingleMark string

	withMonitoring bool

	//"true" or "false"
	Monitoring string

	logger logger.Logger

	withGin bool

	withBeego bool

	withGrpc bool

	writer file_dir.IfaceWriter

	tpl templates.Tpl
}

type ProjectOption func(*Project)

func NewProject(options ...ProjectOption) *Project {

	project := &Project{}

	for _, option := range options {
		option(project)
	}

	project.RunTrans = make([]string, 0)

	project.ImportServer = make([]string, 0)

	project.SingleMark = "`"

	return project
}

func WithProjectLogger(logger logger.Logger) ProjectOption {
	return func(pj *Project) {
		pj.logger = logger
	}
}

func WithProjectWriter(writer file_dir.IfaceWriter) ProjectOption {
	return func(pj *Project) {
		pj.writer = writer
	}
}

func WithProjectTpl(tpl templates.Tpl) ProjectOption {
	return func(pj *Project) {
		pj.tpl = tpl
	}
}

func (pj *Project) Run(v *viper.Viper) {
	pj.bindInput(v)

	pj.getProPath()

	pj.getPackName()

	pj.initTransport()

	pj.createDir()

	pj.build()
}

func (pj *Project) bindInput(v *viper.Viper) bool {
	serverName := v.GetString("server_name")
	if serverName == "" {
		pj.logger.Fatalf("The server_name is empty")
	}
	pj.ServerName = serverName

	if pj.checkServerName() == false {
		pj.logger.Fatalf("The server_name only supports【a-z_-】")
	}

	pj.withGin = v.GetBool("gin")

	pj.withBeego = v.GetBool("beego")

	if pj.withGin == true && pj.withBeego == true {
		pj.logger.Fatalf("either gin or beego")
	}

	if pj.withGin == false && pj.withBeego == false {
		pj.withGin = true
	}

	pj.withGrpc = v.GetBool("grpc")

	pj.withMonitoring = v.GetBool("monitoring")
	if pj.withMonitoring == true {
		pj.Monitoring = "true"
	} else {
		pj.Monitoring = "false"
	}

	return true
}

//checkServerName ServerName only support lowercase ,"_", "-"
func (pj *Project) checkServerName() bool {
	match, err := regexp.MatchString("^[a-z_-]+$", pj.ServerName)
	if err != nil {
		pj.logger.Fatalf(err.Error())
	}

	return match
}

func (pj *Project) createDir() bool {
	exists, err := file_dir.IsExistsDir(pj.ServerName)
	if err != nil {
		pj.logger.Fatalf(err.Error())
	}

	if exists {
		pj.logger.Fatalf("The %s is exists can't be create", pj.ServerName)
	}

	err = file_dir.CreateDir(pj.ServerName)
	if err != nil {
		pj.logger.Fatalf(err.Error())
	}

	return true
}

func (pj *Project) delDir() bool {
	dirExists, err := file_dir.IsExistsDir(pj.ServerName)
	if err != nil {
		pj.logger.Errorf(err.Error())
	}

	if dirExists {
		err = file_dir.RemoveDir(pj.ServerName)
		if err != nil {
			pj.logger.Errorf("remove err : %s", err.Error())
		}
		pj.logger.Infof("remove %s success", pj.ServerName)
	}

	return true
}

func (pj *Project) getProPath() {
	currentDir := file_dir.GetGoProPath()
	if currentDir != "" {
		currentDir = currentDir + string(filepath.Separator)
	}
	pj.ProPath = currentDir
}

//getPackName in most cases,  ServerName eq PackageName
func (pj *Project) getPackName() {
	pj.PackageName = strings.Replace(pj.ServerName, "-", "_", -1)
}

//initTransport initialization Transport mode
func (pj *Project) initTransport() {
	if pj.withGin == true {
		GinInit()
		pj.RunTrans = append(pj.RunTrans, "app.Trans = append(app.Trans, http.NewGinServer(app))")
		pj.ImportServer = append(pj.ImportServer, pj.ProPath+pj.ServerName+"/internal/transports/http")
	}

	if pj.withBeego == true {
		BeegoInit()
		pj.RunTrans = append(pj.RunTrans, "app.Trans = append(app.Trans, http.NewBeegoServer(app.Esim))")
		pj.ImportServer = append(pj.ImportServer, pj.ProPath+pj.ServerName+"/internal/transports/http")
	}

	if pj.withGrpc == true {
		GrpcInit()
		pj.RunTrans = append(pj.RunTrans, "app.Trans = append(app.Trans, grpc.NewGrpcServer(app))")
		pj.ImportServer = append(pj.ImportServer, pj.ProPath+pj.ServerName+"/internal/transports/grpc")
	}
}

//build create a new project locally
//if an error occurred, remove the project
func (pj *Project) build() bool {

	pj.logger.Infof("starting create %s, package name %s", pj.ServerName, pj.PackageName)

	defer func() {
		if err := recover(); err != nil {
			pj.delDir()
		}
	}()

	for _, file := range Files {
		dir := pj.ServerName + string(filepath.Separator) + file.Dir

		exists, err := file_dir.IsExistsDir(dir)
		if err != nil {
			pj.logger.Panicf("%s : %s", file.FileName, err.Error())
		}

		if exists == false {
			err = file_dir.CreateDir(dir)
			if err != nil {
				pj.logger.Panicf("%s : %s", file.FileName, err.Error())
			}
		}

		fileName := dir + string(filepath.Separator) + file.FileName

		content, err := pj.tpl.Execute(file.FileName, file.Content, pj)
		if err != nil {
			pj.logger.Panicf("%s : %s", file.FileName, err.Error())
		}

		var src []byte
		if filepath.Ext(fileName) == ".go" {
			src, err = imports.Process("", []byte(content), nil)
			if err != nil {
				pj.logger.Panicf("%s : %s", file.FileName, err.Error())
			}
		} else {
			src = []byte(content)
		}

		err = pj.writer.Write(fileName, string(src))
		if err != nil {
			pj.logger.Panicf("%s : %s", file.FileName, err.Error())
		}

		pj.logger.Infof("wrote success : %s", fileName)
	}

	pj.logger.Infof("creation complete : %s ", pj.ServerName)
	return true
}

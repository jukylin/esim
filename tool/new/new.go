package new

import (
	"github.com/spf13/viper"
	"github.com/jukylin/esim/pkg/file-dir"
	logger "github.com/jukylin/esim/log"
	"os"
	"strings"
	"golang.org/x/tools/imports"
	"path/filepath"
	"regexp"
	"github.com/jukylin/esim/pkg/templates"
	"sync"
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

func NewProject(logger logger.Logger) *Project {

	project := &Project{}

	project.logger = logger

	project.RunTrans = make([]string, 0)

	project.ImportServer = make([]string, 0)

	project.SingleMark = "`"

	project.tpl = templates.NewTextTpl()

	project.writer = file_dir.NewEsimWriter()

	return project
}

func (pj *Project) Run(v *viper.Viper) {
	pj.bindInput(v)

	pj.getProPath()

	pj.getPackName()

	pj.initTransport()

	pj.createServerDir()

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

func (pj *Project) createServerDir() bool {
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

func (pj *Project) getProPath()  {
	currentDir := file_dir.GetGoProPath()
	if currentDir != ""{
		currentDir = currentDir + string(filepath.Separator)
	}
	pj.ProPath = currentDir
}

//getPackName in most cases,  ServerName eq PackageName
func (pj *Project) getPackName() {
	pj.PackageName = strings.Replace(pj.ServerName, "-", "_", -1)
}

//initTransport initialization Transport mode to do the work
func (pj *Project) initTransport() {
	if pj.withGin == true {
		GinInit()
		pj.RunTrans = append(pj.RunTrans, "app.Trans = append(app.Trans, http.NewGinServer(app))")
		pj.ImportServer = append(pj.ImportServer, pj.ProPath + pj.ServerName + "/internal/transports/http")
	}

	if pj.withBeego == true {
		BeegoInit()
		pj.RunTrans = append(pj.RunTrans, "app.Trans = append(app.Trans, http.NewBeegoServer(app.Esim))")
		pj.ImportServer = append(pj.ImportServer, pj.ProPath + pj.ServerName + "/internal/transports/http")
	}

	if pj.withGrpc == true {
		GrpcInit()
		pj.RunTrans = append(pj.RunTrans, "app.Trans = append(app.Trans, grpc.NewGrpcServer(app))")
		pj.ImportServer = append(pj.ImportServer, pj.ProPath + pj.ServerName + "/internal/transports/grpc")
	}
}

//build create a new project locally
//if an error occurred, remove the project
func (pj *Project) build() bool {

	pj.logger.Infof("starting create %s, package name %s", pj.ServerName, pj.PackageName)

	wg := sync.WaitGroup{}
	wg.Add(len(Files))

	for _, f := range Files {
		go func(file *FileContent) {
			dir := pj.ServerName + string(filepath.Separator) + file.Dir

			exists, err := file_dir.IsExistsDir(dir)
			if err != nil {
				pj.logger.Errorf("%s : %s", file.FileName, err.Error())
				err = os.Remove(pj.ServerName)
				if err != nil {
					pj.logger.Fatalf("remove err : %s", err.Error())
				}

			}

			if exists == false {
				err = file_dir.CreateDir(dir)
				if err != nil {
					pj.logger.Errorf("%s : %s", file.FileName, err.Error())
					err = os.Remove(pj.ServerName)
					if err != nil {
						pj.logger.Fatalf("remove err : %s", err.Error())
					}
				}
			}

			fileName := dir + string(filepath.Separator) + file.FileName

			content, err := pj.tpl.Execute(file.FileName,  file.Content, pj)
			if err != nil {
				pj.logger.Errorf("%s : %s", file.FileName, err.Error())
				err = os.Remove(pj.ServerName)
				if err != nil {
					pj.logger.Fatalf("remove err : %s", err.Error())
				}
			}

			var src []byte
			if filepath.Ext(fileName) == ".go" {
				src, err = imports.Process("", []byte(content), nil)
				if err != nil {
					pj.logger.Errorf("%s : %s", file.FileName, err.Error())
					err = os.Remove(pj.ServerName)
					if err != nil {
						pj.logger.Fatalf("remove err : %s", err.Error())
					}
				}
			}

			err = pj.writer.Write(fileName, string(src))
			if err != nil {
				pj.logger.Errorf("%s : %s", file.FileName, err.Error())
				err = os.Remove(pj.ServerName)
				if err != nil {
					pj.logger.Fatalf("remove err : %s", err.Error())
				}
			}

			pj.logger.Infof("wrote success : %s", fileName)
			wg.Done()
		}(f)
	}

	wg.Wait()
	pj.logger.Infof("creation complete : %s ", pj.ServerName)
	return true
}
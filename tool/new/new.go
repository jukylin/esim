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
	"text/template"
	"bytes"
	"github.com/jukylin/esim/pkg/templates"
	"io/ioutil"
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

	RunServer []string

	ProPath string

	ImportServer []string

	//true or false string type
	Monitoring string

	SingleMark string

	withMonitoring bool

	logger logger.Logger

	withGin bool

	withBeego bool

	withGrpc bool
}

func NewProject(logger logger.Logger) *Project {

	project := &Project{}

	project.logger = logger

	project.RunServer = make([]string, 0)

	project.ImportServer = make([]string, 0)

	project.SingleMark = "`"

	return project
}

func (pj *Project) Run(v *viper.Viper) {
	pj.bindInput(v)

	pj.getProPath()

	pj.getPackName()

	pj.initServer()

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

//PackName In most cases,  ServerName eq PackageName
func (pj *Project) getPackName() {
	pj.PackageName = strings.Replace(pj.ServerName, "-", "_", -1)
}

func (pj *Project) initServer() {
	if pj.withGin == true {
		GinInit()
		pj.RunServer = append(pj.RunServer, "app.Trans = append(app.Trans, http.NewGinServer(app))")
		pj.ImportServer = append(pj.ImportServer, pj.ProPath + pj.ServerName + "/internal/transports/http")
	}

	if pj.withBeego == true {
		BeegoInit()
		pj.RunServer = append(pj.RunServer, "app.Trans = append(app.Trans, http.NewBeegoServer(app.Esim))")
		pj.ImportServer = append(pj.ImportServer, pj.ProPath + pj.ServerName + "/internal/transports/http")
	}

	if pj.withGrpc == true {
		GrpcInit()
		pj.RunServer = append(pj.RunServer, "app.Trans = append(app.Trans, grpc.NewGrpcServer(app))")
		pj.ImportServer = append(pj.ImportServer, pj.ProPath + pj.ServerName + "/internal/transports/grpc")
	}
}

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

			content, err := pj.executeTmpl(file.FileName, file.Content)
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
			}else{
				src = []byte(content)
			}

			err = ioutil.WriteFile(fileName, src, 0666)
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

func (pj *Project) executeTmpl(tplName string, text string) ([]byte, error) {
	tmpl, err := template.New(tplName).Funcs(templates.EsimFuncMap()).
		Parse(text)
	if err != nil{
		return nil, err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, pj)
	if err != nil{
		return nil, err
	}

	return buf.Bytes(), nil
}
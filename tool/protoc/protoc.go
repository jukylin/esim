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
	"errors"
)


type protocer struct {
	target string

	fromProtoc string

	packageName string

	logger logger.Logger
}

type Option func(*protocer)

func NewProtoc(options ...Option) *protocer {
	p := &protocer{}

	for _, option := range options {
		option(p)
	}

	return p
}

func WithProtocLogger(logger logger.Logger) Option {
	return func(p *protocer) {
		p.logger = logger
	}
}

func (p *protocer) Run(v *viper.Viper) bool {
	p.bindInput(v)

	p.logger.Infof("Please confirm that protoc is installed")

	p.execCmd()

	return true
}


func (p *protocer) bindInput(v *viper.Viper) bool {
	target := v.GetString("target")

	ex, err := file_dir.IsExistsDir(target)
	if err != nil {
		p.logger.Fatalf(err.Error())
	}

	if ex == false {
		p.logger.Fatalf("Dir not exists %s", target)
	}
	p.target = target

	fromProto := v.GetString("from_proto")
	if fromProto == "" {
		p.logger.Fatalf("Please special proto file")
	}
	p.fromProtoc = fromProto

	pkgName := v.GetString("package")
	if pkgName == "" {
		pkgName, err = p.parsePkgName(fromProto)
		if err != nil {
			p.logger.Fatalf(err.Error())
		}
	}
	p.packageName = pkgName

	err = file_dir.CreateDir(target + "/" + pkgName)
	if err != nil {
		p.logger.Fatalf("Create fail % : %s", target + string(filepath.Separator) + pkgName, err.Error())
	}

	return true
}

func (p *protocer) execCmd() bool {
	pwd, _ := os.Getwd()

	cmdLine := fmt.Sprintf("protoc --go_out=plugins=grpc:%s %s",
		p.target + string(filepath.Separator) + p.packageName, p.fromProtoc)

	p.logger.Infof(cmdLine)

	args := strings.Split(cmdLine, " ")

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = pwd

	cmd.Env = os.Environ()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		p.logger.Fatalf(err.Error())
	}

	return true
}

//parsePkgName parse the package name from protoc file
//if not found stop the run
func (p *protocer) parsePkgName(protoFile string) (string, error) {

	if filepath.Ext(protoFile) != ".proto" {
		return "", errors.New(fmt.Sprintf("It is not the proto file : %s", protoFile))
	}

	f, err := os.Open(protoFile)
	if err != nil {
		return "", err
	}
	defer f.Close()

	rd := bufio.NewReader(f)
	for {
		line, err := rd.ReadString('\n')

		if err != nil || io.EOF == err {
			break
		}

		match, err := regexp.MatchString("^package", line)
		if err != nil {
			return "", nil
		}

		if match {
			reg, err := regexp.Compile(`\w+`)
			if err != nil {
				return "", err
			}

			strs := reg.FindAllString(line, -1)
			if len(strs) > 1 {
				return strs[1], nil
			}
		}
	}

	return "", errors.New(fmt.Sprintf("Not found the package name from protoc file"))
}

package protoc

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jukylin/esim/log"
	filedir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/spf13/viper"
)

const (
	protocCmd = "protoc"
)

type Protocer struct {
	target string

	fromProto string

	protoPath string

	packageName string

	logger log.Logger
}

type Option func(*Protocer)

func NewProtocer(options ...Option) *Protocer {
	p := &Protocer{}

	for _, option := range options {
		option(p)
	}

	return p
}

func WithProtocLogger(logger log.Logger) Option {
	return func(p *Protocer) {
		p.logger = logger
	}
}

func (p *Protocer) Run(v *viper.Viper) bool {
	p.bindInput(v)

	p.logger.Infof("Please confirm that protoc is installed")

	p.parseProtoPath()

	p.execCmd()

	return true
}

func (p *Protocer) bindInput(v *viper.Viper) bool {
	target := v.GetString("target")

	ex, err := filedir.IsExistsDir(target)
	if err != nil {
		p.logger.Fatalf(err.Error())
	}

	if !ex {
		p.logger.Fatalf("Dir not exists %s", target)
	}
	if target != string(filepath.Separator) {
		p.target = strings.TrimRight(target, string(filepath.Separator))
	} else {
		p.target = target
	}

	fromProto := v.GetString("from_proto")
	if fromProto == "" {
		p.logger.Fatalf("Please special proto file")
	}
	p.fromProto = fromProto

	pkgName := v.GetString("package")
	if pkgName == "" {
		pkgName, err = p.parsePkgName(fromProto)
		if err != nil {
			p.logger.Fatalf(err.Error())
		}
	}
	p.packageName = pkgName

	err = filedir.CreateDir(target + string(filepath.Separator) + pkgName)
	if err != nil {
		p.logger.Fatalf("Create fail % : %s", target+string(filepath.Separator)+pkgName, err.Error())
	}

	return true
}

func (p *Protocer) parseProtoPath() {
	strs := strings.Split(p.fromProto, string(filepath.Separator))
	protoPath := strs[0 : len(strs)-1]
	p.protoPath = strings.Join(protoPath, string(filepath.Separator))
}

func (p *Protocer) execCmd() bool {
	pwd, err := os.Getwd()
	if err != nil {
		p.logger.Fatalf(err.Error())
	}

	cmdLine := fmt.Sprintf("--go_out=plugins=grpc:%s --proto_path %s %s",
		p.target+string(filepath.Separator)+p.packageName, p.protoPath, p.fromProto)

	p.logger.Infof("%s %s", protocCmd, cmdLine)

	args := strings.Split(cmdLine, " ")

	cmd := exec.Command(protocCmd, args...)
	cmd.Dir = pwd

	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		p.logger.Fatalf(err.Error())
	}

	return true
}

// 从protoc文件，解析包名，如果没有找到，停止运行.
func (p *Protocer) parsePkgName(protoFile string) (string, error) {
	if filepath.Ext(protoFile) != ".proto" {
		return "", fmt.Errorf("it is not the proto file : %s", protoFile)
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
		reg := regexp.MustCompile("^package")

		if reg.FindString(line) != "" {
			reg := regexp.MustCompile(`\w+`)
			if err != nil {
				return "", err
			}

			strs := reg.FindAllString(line, -1)
			if len(strs) > 1 {
				return strs[1], nil
			}
		}
	}

	return "", fmt.Errorf("not found the package name from protoc file")
}

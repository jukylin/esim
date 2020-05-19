package db2entity

import (
	"strings"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	file_dir "github.com/jukylin/esim/pkg/file-dir"

	"errors"
	"os"
	"path/filepath"

	"github.com/jukylin/esim/infra"
	"github.com/jukylin/esim/pkg/templates"
	domain_file "github.com/jukylin/esim/tool/db2entity/domain-file"
	"github.com/serenize/snaker"
	"github.com/spf13/viper"
	"golang.org/x/tools/imports"
)

type Db2Entity struct {
	// Camel Form
	CamelStruct string

	ColumnsRepo domain_file.ColumnsRepo

	DbConf *domain_file.DbConfig

	writer file_dir.IfaceWriter

	execer pkg.Exec

	domainFiles []domain_file.DomainFile

	// parsed content
	domainContent map[string]string

	// record wrote content, if an error occurred rollback the file
	wroteContent map[string]string

	tpl templates.Tpl

	shareInfo *domain_file.ShareInfo

	injectInfos []*domain_file.InjectInfo

	infraer *infra.Infraer

	// if true inject repo to infra
	withInject bool

	logger log.Logger

	withPackage string

	withStruct string

	withInfraDir string

	withInfraFile string
}

type Db2EnOption func(*Db2Entity)

type Db2EnOptions struct{}

func NewDb2Entity(options ...Db2EnOption) *Db2Entity {
	d := &Db2Entity{}

	for _, option := range options {
		option(d)
	}

	if d.writer == nil {
		d.writer = file_dir.NewNullWrite()
	}

	if d.execer == nil {
		d.execer = &pkg.NullExec{}
	}

	if d.logger == nil {
		d.logger = log.NewNullLogger()
	}

	d.domainContent = make(map[string]string)
	d.wroteContent = make(map[string]string)

	return d
}

func (Db2EnOptions) WithLogger(logger log.Logger) Db2EnOption {
	return func(d *Db2Entity) {
		d.logger = logger
	}
}

func (Db2EnOptions) WithColumnsInter(columnsRepo domain_file.ColumnsRepo) Db2EnOption {
	return func(d *Db2Entity) {
		d.ColumnsRepo = columnsRepo
	}
}

func (Db2EnOptions) WithWriter(writer file_dir.IfaceWriter) Db2EnOption {
	return func(d *Db2Entity) {
		d.writer = writer
	}
}

func (Db2EnOptions) WithExecer(execer pkg.Exec) Db2EnOption {
	return func(d *Db2Entity) {
		d.execer = execer
	}
}

func (Db2EnOptions) WithDbConf(dbConf *domain_file.DbConfig) Db2EnOption {
	return func(d *Db2Entity) {
		d.DbConf = dbConf
	}
}

func (Db2EnOptions) WithDomainFile(dfs ...domain_file.DomainFile) Db2EnOption {
	return func(d *Db2Entity) {
		d.domainFiles = dfs
	}
}

func (Db2EnOptions) WithShareInfo(shareInfo *domain_file.ShareInfo) Db2EnOption {
	return func(d *Db2Entity) {
		d.shareInfo = shareInfo
	}
}

func (Db2EnOptions) WithTpl(tpl templates.Tpl) Db2EnOption {
	return func(d *Db2Entity) {
		d.tpl = tpl
	}
}

func (Db2EnOptions) WithInfraer(infraer *infra.Infraer) Db2EnOption {
	return func(d *Db2Entity) {
		d.infraer = infraer
	}
}

func (de *Db2Entity) Run(v *viper.Viper) error {
	defer func() {
		if err := recover(); err != nil {
			de.logger.Errorf("A panic occurred : %s", err)

			if len(de.wroteContent) > 0 {
				for path := range de.wroteContent {
					de.logger.Debugf("remove %s", path)
					err := os.RemoveAll(path)
					if err != nil {
						de.logger.Errorf(err.Error())
					}
				}
			}
		}
	}()

	de.bindInput(v)

	de.shareInfo.CamelStruct = de.CamelStruct

	if len(de.domainFiles) < 1 {
		return errors.New("have not domain file")
	}

	//select table's columns from repository
	cs, err := de.ColumnsRepo.SelectColumns(de.DbConf)
	if err != nil {
		return err
	}

	if !cs.IsEntity() {
		return errors.New("it is not the entity")
	}

	err = de.generateDomainFile(v, cs)
	if err != nil {
		return err
	}

	//save domain content
	if len(de.domainContent) > 0 {
		for path, content := range de.domainContent {
			de.logger.Debugf("writing %s", path)
			err = de.writer.Write(path, content)
			if err != nil {
				de.logger.Panicf(err.Error())
			}
			de.wroteContent[path] = content
		}
	}

	de.injectToInfra(v)

	return nil
}

func (de *Db2Entity) bindInput(v *viper.Viper) {
	de.DbConf.ParseConfig(v, de.logger)

	packageName := v.GetString("package")
	if packageName == "" {
		packageName = de.DbConf.Database
	}
	de.withPackage = packageName

	stuctName := v.GetString("struct")
	if stuctName == "" {
		stuctName = de.DbConf.Table
	}
	de.withStruct = stuctName
	de.CamelStruct = snaker.SnakeToCamel(stuctName)

	de.logger.Debugf("CamelStruct : %s", de.CamelStruct)

	de.bindInfra(v)
}

func (de *Db2Entity) bindInfra(v *viper.Viper) {
	de.withInject = v.GetBool("inject")

	de.withInfraDir = v.GetString("infra_dir")
	if de.withInfraDir == "" {
		de.withInfraDir = "internal" + string(filepath.Separator) + "infra" +
			string(filepath.Separator)
	} else {
		de.withInfraDir = strings.TrimLeft(de.withInfraDir, ".") + string(filepath.Separator)
		de.withInfraDir = strings.Trim(de.withInfraDir, "/") + string(filepath.Separator)
	}

	if v.GetString("infra_file") == "" {
		de.withInfraFile = "infra.go"
	}

	exists, err := file_dir.IsExistsFile(de.withInfraDir + de.withInfraFile)
	if err != nil {
		de.logger.Fatalf(err.Error())
		return
	}

	if !exists {
		de.logger.Fatalf("%s not exists", de.withInfraDir+de.withInfraFile)
	}
}

//injectToInfra inject repo to infra.go and execute wire command
func (de *Db2Entity) injectToInfra(v *viper.Viper) {
	if !de.withInject {
		de.logger.Infof("disable inject")
		return
	}

	de.infraer.Inject(v, de.injectInfos)
}

func (de *Db2Entity) makeCodeBeautiful(src string) string {
	result, err := imports.Process("", []byte(src), nil)
	if err != nil {
		de.logger.Panicf("err %s : %s", err.Error(), src)
		return ""
	}

	return string(result)
}

func (de *Db2Entity) generateDomainFile(v *viper.Viper, cs domain_file.Columns) error {
	var content string

	//loop domainFiles to generate domain file
	for _, df := range de.domainFiles {
		err := df.BindInput(v)
		if err != nil {
			return err
		}

		if !df.Disabled() {

			de.shareInfo.ParseInfo(df)

			df.ParseCloumns(cs, de.shareInfo)

			//parsed template
			content = df.Execute()

			content = de.makeCodeBeautiful(content)

			de.domainContent[df.GetSavePath()] = content

			injectInfo := df.GetInjectInfo()

			if injectInfo != nil {
				de.injectInfos = append(de.injectInfos, injectInfo)
			}
		} else {
			de.logger.Infof("disabled %s", df.GetName())
		}
	}

	return nil
}

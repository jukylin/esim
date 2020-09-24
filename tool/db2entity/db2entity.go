package db2entity

import (
	"strings"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	filedir "github.com/jukylin/esim/pkg/file-dir"

	"errors"
	"os"
	"path/filepath"

	"github.com/jukylin/esim/infra"
	"github.com/jukylin/esim/pkg/templates"
	domainfile "github.com/jukylin/esim/tool/db2entity/domain-file"
	"github.com/serenize/snaker"
	"github.com/spf13/viper"
	"golang.org/x/tools/imports"
)

type Db2Entity struct {
	// 驼峰格式
	CamelStruct string

	ColumnsRepo domainfile.ColumnsRepo

	DbConf *domainfile.DbConfig

	writer filedir.IfaceWriter

	execer pkg.Exec

	domainFiles []domainfile.DomainFile

	// 记录解析的内容
	domainContent map[string]string

	// 用于记录已经写入本地的文件，如果发生错误回滚
	wroteContent map[string]string

	tpl templates.Tpl

	shareInfo *domainfile.ShareInfo

	injectInfos []*domainfile.InjectInfo

	infraer *infra.Infraer

	// 把资源注入到基础设施
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
		d.writer = filedir.NewNullWrite()
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

func (Db2EnOptions) WithColumnsInter(columnsRepo domainfile.ColumnsRepo) Db2EnOption {
	return func(d *Db2Entity) {
		d.ColumnsRepo = columnsRepo
	}
}

func (Db2EnOptions) WithWriter(writer filedir.IfaceWriter) Db2EnOption {
	return func(d *Db2Entity) {
		d.writer = writer
	}
}

func (Db2EnOptions) WithExecer(execer pkg.Exec) Db2EnOption {
	return func(d *Db2Entity) {
		d.execer = execer
	}
}

func (Db2EnOptions) WithDbConf(dbConf *domainfile.DbConfig) Db2EnOption {
	return func(d *Db2Entity) {
		d.DbConf = dbConf
	}
}

func (Db2EnOptions) WithDomainFile(dfs ...domainfile.DomainFile) Db2EnOption {
	return func(d *Db2Entity) {
		d.domainFiles = dfs
	}
}

func (Db2EnOptions) WithShareInfo(shareInfo *domainfile.ShareInfo) Db2EnOption {
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

	if len(de.domainFiles) == 0 {
		return errors.New("没有找到领域文件")
	}

	// 从仓库中读取表字段信息
	cs, err := de.ColumnsRepo.SelectColumns(de.DbConf)
	if err != nil {
		return err
	}

	if !cs.IsEntity() {
		return errors.New("非实体")
	}

	err = de.generateDomainFile(v, cs)
	if err != nil {
		return err
	}

	// 保存领域内容
	if len(de.domainContent) > 0 {
		for path, content := range de.domainContent {
			de.logger.Debugf("正在写入 %s", path)
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

	de.logger.Debugf("驼峰法 : %s", de.CamelStruct)

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

	exists, err := filedir.IsExistsFile(de.withInfraDir + de.withInfraFile)
	if err != nil {
		de.logger.Fatalf(err.Error())
		return
	}

	if !exists {
		de.logger.Fatalf("%s 不存在", de.withInfraDir+de.withInfraFile)
	}
}

// injectToInfra 把资源注入到基础设施，并运行wire
func (de *Db2Entity) injectToInfra(v *viper.Viper) {
	if !de.withInject {
		de.logger.Infof("自动注入被关闭")
		return
	}

	de.infraer.Inject(v, de.injectInfos)
}

func (de *Db2Entity) makeCodeBeautiful(src string) string {
	options := &imports.Options{}
	options.Comments = true
	options.TabIndent = true
	options.TabWidth = 8
	options.FormatOnly = true

	result, err := imports.Process("", []byte(src), options)
	if err != nil {
		de.logger.Panicf("错误 %s : %s", err.Error(), src)
		return ""
	}

	return string(result)
}

func (de *Db2Entity) generateDomainFile(v *viper.Viper, cs domainfile.Columns) error {
	var content string

	// 生成领域文件
	for _, df := range de.domainFiles {
		err := df.BindInput(v)
		if err != nil {
			return err
		}

		if !df.Disabled() {
			de.shareInfo.ParseInfo(df)

			df.ParseCloumns(cs, de.shareInfo)

			// 解析模板
			content = df.Execute()

			content = de.makeCodeBeautiful(content)

			de.domainContent[df.GetSavePath()] = content

			injectInfo := df.GetInjectInfo()

			if injectInfo != nil {
				de.injectInfos = append(de.injectInfos, injectInfo)
			}
		} else {
			de.logger.Infof("关闭 %s", df.GetName())
		}
	}

	return nil
}

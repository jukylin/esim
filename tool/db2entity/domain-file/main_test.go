package domainfile

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg/templates"
	"github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"
)

var (
	testEntityDomainFile DomainFile

	testDaoDomainFile DomainFile

	testRepoDomainFile DomainFile

	db *sql.DB
)

func TestMain(m *testing.M) {
	loggerOptions := log.LoggerOptions{}
	logger := log.NewLogger(loggerOptions.WithDebug(true))

	tpl := templates.NewTextTpl()

	testEntityDomainFile = NewEntityDomainFile(
		WithEntityDomainFileLogger(logger),
		WithEntityDomainFileTpl(tpl))

	testDaoDomainFile = NewDaoDomainFile(
		WithDaoDomainFileLogger(logger),
		WithDaoDomainFileTpl(tpl))

	testRepoDomainFile = NewRepoDomainFile(
		WithRepoDomainFileLogger(logger),
		WithRepoDomainFileTpl(tpl))

	pool, err := dockertest.NewPool("")
	if err != nil {
		logger.Fatalf("Could not connect to docker: %s", err)
	}

	opt := &dockertest.RunOptions{
		Repository: "mysql",
		Tag:        "latest",
		Env:        []string{"MYSQL_ROOT_PASSWORD=123456"},
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(opt, func(hostConfig *dc.HostConfig) {
		hostConfig.PortBindings = map[dc.Port][]dc.PortBinding{
			"3306/tcp": {{HostIP: "", HostPort: "3306"}},
		}
	})
	if err != nil {
		logger.Fatalf("Could not start resource: %s", err)
	}

	err = resource.Expire(100)
	if err != nil {
		logger.Fatalf(err.Error())
	}

	if err := pool.Retry(func() error {
		var err error
		db, err = sql.Open("mysql",
			"root:123456@tcp(localhost:3306)/mysql?charset=utf8&parseTime=True&loc=Local")
		if err != nil {
			return err
		}
		db.SetMaxOpenConns(100)

		return db.Ping()
	}); err != nil {
		logger.Fatalf("Could not connect to docker: %s", err)
	}

	sqls := []string{
		`create database test_1;`,
		`CREATE TABLE IF NOT EXISTS test_1.test(
		  id int not NULL auto_increment,
		  title VARCHAR(10) not NULL DEFAULT '',
		  PRIMARY KEY (id)
		)engine=innodb;`}

	for _, execSQL := range sqls {
		res, err := db.Exec(execSQL)
		if err != nil {
			logger.Errorf(err.Error())
		}
		_, err = res.RowsAffected()
		if err != nil {
			logger.Errorf(err.Error())
		}
	}

	code := m.Run()

	db.Close()
	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		logger.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

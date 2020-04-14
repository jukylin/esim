package file_dir

import (
	"os"
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"
	"github.com/jukylin/esim/log"
)


func CreateFile(file string) (bool, error) {
	f, err := os.Create(file)
	if err != nil{
		return false, err
	}else{
		f.Close()
		return true, nil
	}
}


func IsExistsDir(dir string)  (bool, error) {
	_, err := os.Stat(dir)
	if err == nil{
		return true, err
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func CreateDir(dir string) error {
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil{
		return err
	}

	return nil
}


func IsExistsFile(file string) (bool, error) {
	if _, err := os.Stat(file); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

func IsEmptyDir(dir string) (bool, error) {

	exists, err := IsExistsDir(dir)
	if err != nil{
		return false, err
	}

	if exists == false{
		return false, errors.New("目录不存在")
	}


	dirs, err := ioutil.ReadDir(dir)

	if err != nil{
		return false, err
	}

	if len(dirs) == 0{
		return true, nil
	}else {
		return false, nil
	}
}

func GetParDir() string {
	wd,err := os.Getwd()
	if err != nil {
		panic(err)
	}
	path := filepath.Dir(wd) + string(filepath.Separator)
	parDir := strings.Replace(wd, path, "", -1)

	return parDir
}


func GetCurrentDir() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	wd = strings.TrimRight(wd, string(filepath.Separator))

	return wd
}


func GetGoProPath() string {
	wd,err := os.Getwd()
	if err != nil {
		panic(err)
	}

	gopath := os.Getenv("GOPATH")
	if gopath == ""{
		panic(errors.New("not set GOPATH"))
	}

	srcPath := gopath + string(filepath.Separator) + "src" + string(filepath.Separator)

	parDir := strings.Replace(wd + string(filepath.Separator), srcPath, "", -1)

	parDir = strings.Trim(parDir, string(filepath.Separator))

	return parDir
}

func RemoveDir(dir string) error {
	return os.RemoveAll(dir)
}

//BackUpFile backup files to os.Getenv("GOPATH") + "/pkg/esim/backup/"
//backFile is Absolute path
// Overwrite as soon as the file exists
func EsimBackUpFile(backFile string) error {

	if backFile == ""{
		return errors.New("没有文件")
	}

	dir := filepath.Dir(backFile)
	relativeDir := strings.Replace(dir, os.Getenv("GOPATH") + string(filepath.Separator) +
		"src" + string(filepath.Separator), "", -1)

	backUpPath := os.Getenv("GOPATH") + string(filepath.Separator) + "pkg" +
		string(filepath.Separator) + "esim" + string(filepath.Separator) + "backup" + string(filepath.Separator)
	targetPath := backUpPath + relativeDir
	exists, err := IsExistsDir(targetPath)
	if err != nil{
		return err
	}

	if exists == false {
		err = CreateDir(targetPath)
		if err != nil {
			return err
		}
	}

	relativePath := strings.Replace(backFile, os.Getenv("GOPATH") + string(filepath.Separator) +
		"src" + string(filepath.Separator), "", -1)
	fileExists, err := IsExistsFile(backUpPath + relativePath)
	if err != nil{
		return err
	}

	if fileExists == false {
		_, err = CreateFile(backUpPath + relativePath)
		if err != nil {
			return err
		}
	}

	input, err := ioutil.ReadFile(backFile)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(backUpPath + relativePath, input, 0644)
	if err != nil {
		return err
	}

	log.Log.Infof("%s backup to %s", relativePath, backUpPath)

	return nil
}

//EsimRecoverFile recover file from os.Getenv("GOPATH") + "/pkg/esim/backup/"
func EsimRecoverFile(recoverFile string) error {

	if recoverFile == ""{
		return errors.New("没有文件")
	}

	relativeFile := strings.Replace(recoverFile, os.Getenv("GOPATH") + string(filepath.Separator) +
		"src" + string(filepath.Separator), "", -1)

	backUpPath := os.Getenv("GOPATH") + string(filepath.Separator) + "pkg" + string(filepath.Separator) +
		"esim" + string(filepath.Separator) + "backup" + string(filepath.Separator)
	targetPath := backUpPath + relativeFile
	exists, err := IsExistsDir(targetPath)
	if err != nil {
		return err
	}

	if exists == false {
		return errors.New(targetPath + " not exists")
	}


	fileExists, err := IsExistsFile(recoverFile)
	if err != nil{
		return err
	}

	if fileExists == false {
		_, err = CreateFile(recoverFile)
		if err != nil {
			return err
		}
	}

	input, err := ioutil.ReadFile(targetPath)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(recoverFile, input, 0644)
	if err != nil {
		return err
	}

	log.Log.Infof("%s recover success", recoverFile)

	return nil
}


func EsimWrite(filePath string, content string) error {

	dir := filepath.Dir(filePath)

	exists, err := IsExistsDir(dir)
	if err != nil{
		return err
	}

	if exists == false {
		err = CreateDir(dir)
		if err != nil{
			return err
		}
	}

	dst, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer dst.Close()

	dst.Write([]byte(content))

	return nil
}
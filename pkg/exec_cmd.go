package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)


type Exec interface {
	ExecWire(string) error

	ExecFmt(string) error
}


type CmdExec struct{}

func NewCmdExec() Exec {
	return &CmdExec{}
}

type NullExec struct{}

func NewNullExec() Exec {
	return &NullExec{}
}

func (this *CmdExec) ExecWire(dir string) error {
	cmd_line := fmt.Sprintf("wire")

	args := strings.Split(cmd_line, " ")

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir

	cmd.Env = os.Environ()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	return err
}


func (this *CmdExec) ExecFmt(dir string) error {
	return nil
}



func (this *NullExec) ExecWire(dir string) error {
	return nil
}


func (this *NullExec) ExecFmt(dir string) error {
	return nil
}
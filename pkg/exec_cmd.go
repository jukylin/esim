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

func (ce *CmdExec) ExecWire(dir string) error {
	cmdLine := fmt.Sprintf("wire")

	args := strings.Split(cmdLine, " ")

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir

	cmd.Env = os.Environ()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	return err
}

func (ce *CmdExec) ExecFmt(dir string) error {
	return nil
}

func (ce *NullExec) ExecWire(dir string) error {
	return nil
}

func (ce *NullExec) ExecFmt(dir string) error {
	return nil
}

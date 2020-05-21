package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Exec interface {
	ExecWire(string, ...string) error

	ExecFmt(string, ...string) error
}

type CmdExec struct{}

func NewCmdExec() Exec {
	return &CmdExec{}
}

func (ce *CmdExec) ExecWire(dir string, args ...string) error {
	cmd := exec.Command("wire", args...)
	cmd.Dir = dir

	cmd.Env = os.Environ()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	return err
}

func (ce *CmdExec) ExecFmt(dir string, args ...string) error {
	return nil
}

func (ce *CmdExec) ExecBuild(dir string, args ...string) error {
	cmdLine := fmt.Sprintf("build")

	args = append(strings.Split(cmdLine, " "), args...)

	cmd := exec.Command("go", args...)
	cmd.Dir = dir

	cmd.Env = os.Environ()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	return err
}

type NullExec struct{}

func NewNullExec() Exec {
	return &NullExec{}
}

func (ce *NullExec) ExecWire(dir string, args ...string) error {
	return nil
}

func (ce *NullExec) ExecFmt(dir string, args ...string) error {
	return nil
}

func (ce *NullExec) ExecBuild(dir string, args ...string) error {
	return nil
}

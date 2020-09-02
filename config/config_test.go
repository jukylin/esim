package config

import (
	"testing"
)

func TestNewViperConfig(t *testing.T) {
	options := ViperConfOptions{}

	file := []string{"./a.yaml", "b.yaml", "c.yaml"}

	conf := NewViperConfig(options.WithConfigType("yaml"),
		options.WithConfFile(file))

	name := conf.GetString("name")
	if name != "esim" {
		t.Errorf("error should esim , now %s", name)
	}

	version := conf.GetFloat64("version")
	if version != 1.0 {
		t.Errorf("error should 1.0 , now %f", version)
	}

	disable := conf.GetBool("disable")
	if disable != true {
		t.Errorf("error should true , now false")
	}
}

func TestNotConfigFile(t *testing.T) {
	options := ViperConfOptions{}

	NewViperConfig(options.WithConfigType("yaml"))
}

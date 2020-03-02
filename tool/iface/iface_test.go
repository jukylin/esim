package iface

import (
	"testing"
)

func TestIface_FindIface(t *testing.T) {
	iface := &Iface{}

	iface.StructName = "TestStub"

	ifacePath := "./example"

	iface.FindIface(ifacePath, "Test")

	src := iface.Gen()


}
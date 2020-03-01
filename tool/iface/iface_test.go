package iface

import "testing"

func TestIface_FindIface(t *testing.T) {
	iface := &Iface{}

	ifacePath := "./example"

	iface.FindIface(ifacePath, "Test")
}
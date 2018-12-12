package controller

import (
	"github.com/kubedge/kubedge-operator-go/pkg/controller/kubedge"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, kubedge.Add)
}

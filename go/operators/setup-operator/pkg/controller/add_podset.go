package controller

import (
	"home/centos/go/operators/setup-operator/pkg/controller/podset"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, podset.Add)
}

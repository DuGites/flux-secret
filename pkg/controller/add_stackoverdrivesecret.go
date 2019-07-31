package controller

import (
	"github.com/flux-secret/pkg/controller/stackoverdrivesecret"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, stackoverdrivesecret.Add)
}

package controller

import (
	"github.com/infracloudio/work8spaces/pkg/controller/workspace"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, workspace.Add)
}

package fakearray

import "github.com/sanderdescamps/go-purefa-mock/pkg/flashclient"

func (array *Array) GetControllers() []flashclient.Controller {
	controllers := make([]flashclient.Controller, len(array.Controllers))
	for i, controller := range array.Controllers {
		controllers[i] = *controller
	}
	return controllers
}

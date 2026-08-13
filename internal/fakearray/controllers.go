package fakearray

import faclient "github.com/sanderdescamps/go-purefa"

func (array *Array) GetControllers() []faclient.Controller {
	controllers := make([]faclient.Controller, len(array.Controllers))
	for i, controller := range array.Controllers {
		controllers[i] = *controller
	}
	return controllers
}

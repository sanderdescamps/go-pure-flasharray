package fakearray

import faclient "github.com/sanderdescamps/go-purefa"

func (array *Array) GetHardware(filters ...func(*faclient.Hardware) bool) []faclient.Hardware {
	hardware := []faclient.Hardware{}
	for _, hw := range array.Hardware {
		matches := true
		for _, filter := range filters {
			if !filter(hw) {
				matches = false
				break
			}
		}
		if matches {
			hardware = append(hardware, *hw)
		}
	}
	return hardware
}

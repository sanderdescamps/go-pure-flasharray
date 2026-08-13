package fakearray

import "github.com/sanderdescamps/go-pure-flasharray/pkg/flashclient"

func (array *Array) GetHardware(filters ...func(*flashclient.Hardware) bool) []flashclient.Hardware {
	hardware := []flashclient.Hardware{}
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

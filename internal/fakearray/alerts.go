package fakearray

import (
	"github.com/sanderdescamps/go-pure-flasharray/pkg/flashclient"
)

func AlertsWithFlagged(flagged bool) func(*flashclient.Alert) bool {
	return func(a *flashclient.Alert) bool {
		return a.Flagged == flagged
	}
}

func (array *Array) GetAlerts(filters ...func(*flashclient.Alert) bool) ([]flashclient.Alert, error) {
	alerts := []flashclient.Alert{}
	for _, alert := range array.Alerts {
		matches := true
		for _, filter := range filters {
			if !filter(alert) {
				matches = false
				break
			}
		}
		if matches {
			alerts = append(alerts, *alert)
		}
	}
	return alerts, nil
}

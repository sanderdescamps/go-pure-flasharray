package fakearray

import faclient "github.com/sanderdescamps/go-purefa"

func AlertsWithFlagged(flagged bool) func(*faclient.Alert) bool {
	return func(a *faclient.Alert) bool {
		return a.Flagged == flagged
	}
}

func (array *Array) GetAlerts(filters ...func(*faclient.Alert) bool) ([]faclient.Alert, error) {
	alerts := []faclient.Alert{}
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

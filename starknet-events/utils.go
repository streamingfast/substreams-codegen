package starknet_events

import (
	"strings"
)

func setNonGoldenAliases(potentialsGoldenEvent map[string]*StarknetEvent, goldenName string, aliases []*Alias) []*Alias {
	for _, event := range potentialsGoldenEvent {
		eventName := event.Name

		if eventName == goldenName {
			continue
		}

		_, newName := eventNameInfo(eventName)

		alias := NewAlias(event.Name, newName)
		aliases = append(aliases, alias)
	}

	return aliases
}

func eventNameInfo(eventName string) (lastPart, aliasName string) {
	splitEventName := strings.Split(eventName, "::")
	if len(splitEventName) < 2 {
		panic("parsed event name does not contain enough parts to have an alias")
	}

	lastPart = splitEventName[len(splitEventName)-1]
	return lastPart, splitEventName[len(splitEventName)-2] + lastPart
}

func detectGoldenEvent(potentialsGoldenEvent map[string]*StarknetEvent) string {
	for _, event := range potentialsGoldenEvent {
		seen := make(map[string]struct{})
		for _, variant := range event.Variants {
			if _, found := potentialsGoldenEvent[variant.Type]; found {
				seen[variant.Type] = struct{}{}
			}

			// Equivalent: Current Event Enum contains all other potentials "golden" events
			if len(seen) == (len(potentialsGoldenEvent) - 1) {
				return event.Name
			}
		}
	}

	return ""
}

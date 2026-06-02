package utils

import "github.com/wisemonkeys-co/event-mapping/types"

type DbInterface interface {
	FindRealmEventMappingData(realmName, event string) (eventMappingData types.Event, location string, err error)
}

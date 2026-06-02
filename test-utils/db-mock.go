package testutils

import "github.com/wisemonkeys-co/event-mapping/types"

type DBMock struct {
	LocationToReturn         string
	EventMappingDataToReturn types.Event
	ErrorToReturn            error
}

func (dm *DBMock) FindRealmEventMappingData(realmName, event string) (eventMappingData types.Event, location string, err error) {
	return dm.EventMappingDataToReturn, dm.LocationToReturn, dm.ErrorToReturn
}

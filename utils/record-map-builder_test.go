package utils

import (
	"fmt"
	"testing"

	testutils "github.com/wisemonkeys-co/event-mapping/test-utils"
	"github.com/wisemonkeys-co/event-mapping/types"
)

func TestSelectNeastedArrayAddress(t *testing.T) {
	testCases := []struct {
		description      string
		previousAddress  string
		canditadeAddress string
		selectedAddress  string
		err              error
	}{
		{
			description:      "empty previous and candidate address",
			previousAddress:  "",
			canditadeAddress: "",
			selectedAddress:  "",
			err:              nil,
		},
		{
			description:      "valid candidate address with one level",
			previousAddress:  "",
			canditadeAddress: "foo.bar.$.sunda.bobra",
			selectedAddress:  "foo.bar",
			err:              nil,
		},
		{
			description:      "valid candidate address with two levels",
			previousAddress:  "",
			canditadeAddress: "foo.bar.$.sunda.bobra.$.akaxike",
			selectedAddress:  "foo.bar.$.sunda.bobra",
			err:              nil,
		},
		{
			description:      "valid candidate address with two levels and a previous with one level",
			previousAddress:  "foo.bar",
			canditadeAddress: "foo.bar.$.sunda.bobra.$.akaxike",
			selectedAddress:  "foo.bar.$.sunda.bobra",
			err:              nil,
		},
		{
			description:      "invalid candidate address",
			previousAddress:  "foo.bar",
			canditadeAddress: "foo.azuna.$.id",
			selectedAddress:  "",
			err: fmt.Errorf(
				"multiple address branches not allowed (%s | %s)",
				"foo.bar",
				"foo.azuna",
			),
		},
		{
			description:      "valid candidate address with one level and previous address with two levels",
			previousAddress:  "foo.bar.$.sunda.bobra",
			canditadeAddress: "foo.bar.$.id",
			selectedAddress:  "foo.bar.$.sunda.bobra",
			err:              nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			result, err := selectNeastedArrayAddress(testCase.previousAddress, testCase.canditadeAddress)
			if err != nil {
				testutils.AssertEquals(t, err.Error(), testCase.err.Error())
			}
			testutils.AssertEquals(t, result, testCase.selectedAddress)
		})
	}
}

func TestGetRecordMapForMapRecordSuccess(t *testing.T) {
	jsonFiles, err := testutils.ReadJsonFiles()
	if err != nil {
		t.Error(err)
		return
	}
	realms := make([]types.Realm, 0)
	realmItems := jsonFiles["realms"]
	for _, realmItem := range realmItems {
		realm, _ := realmItem.(types.Realm)
		realms = append(realms, realm)
	}
	dbMock := &testutils.DBMock{
		LocationToReturn:         realms[0].Location,
		EventMappingDataToReturn: realms[0].Events[0],
	}
	recordMap, eventConfig, err := GetRecordMap(dbMock, realms[0].Name, realms[0].Events[0].Name)
	if err != nil {
		t.Error(t)
		return
	}
	testutils.AssertEquals(t, recordMap.Realm, realms[0].Name)
	testutils.AssertEquals(t, recordMap.Event, realms[0].Events[0].Name)
	testutils.AssertEquals(t, recordMap.Id[0].Name, realms[0].Events[0].FieldMapping[0].Name)
	testutils.AssertEquals(t, recordMap.Id[0].Index, realms[0].Events[0].FieldMapping[0].Index)
	testutils.AssertEquals(t, recordMap.Id[0].RemoteName, realms[0].Events[0].FieldMapping[0].RemoteName)
	testutils.AssertEquals(t, recordMap.Date.Name, realms[0].Events[0].FieldMapping[1].Name)
	testutils.AssertEquals(t, recordMap.Date.Index, realms[0].Events[0].FieldMapping[1].Index)
	testutils.AssertEquals(t, recordMap.Date.Index, realms[0].Events[0].FieldMapping[1].Index)
	testutils.AssertEquals(t, fmt.Sprintf("%+v", eventConfig), fmt.Sprintf("%+v", realms[0].Events[0]))
}

func TestGetRecordMapForMapRecordWithSplitEventFieldSuccess(t *testing.T) {
	realmName := "bonuz-br"
	eventName := "test-value-variables"
	eventData, eventLocation, err := testutils.GetEventDataFromJsonFiles(eventName, realmName)
	if err != nil {
		t.Error(err)
		return
	}
	dbMock := &testutils.DBMock{
		LocationToReturn:         eventLocation,
		EventMappingDataToReturn: eventData,
	}
	recordMap, _, err := GetRecordMap(dbMock, realmName, eventName)
	if err != nil {
		t.Error(t)
		return
	}
	testutils.AssertEquals(t, recordMap.Realm, realmName)
	testutils.AssertEquals(t, recordMap.Event, eventName)
	testutils.AssertEquals(t, recordMap.SplitEventField, "contracts")
	testutils.AssertEquals(t, len(recordMap.Id), 1)
	testutils.AssertEquals(t, recordMap.Id[0].Unique, true)
}

func TestGetRecordMapForMapRecordWithSplitEventFieldNeastedArraySuccess(t *testing.T) {
	realmName := "bonuz-br"
	eventName := "test-split-neasted-array-with-variables"
	eventData, eventLocation, err := testutils.GetEventDataFromJsonFiles(eventName, realmName)
	if err != nil {
		t.Error(err)
		return
	}
	dbMock := &testutils.DBMock{
		LocationToReturn:         eventLocation,
		EventMappingDataToReturn: eventData,
	}
	recordMap, _, err := GetRecordMap(dbMock, realmName, eventName)
	if err != nil {
		t.Error(t)
		return
	}
	testutils.AssertEquals(t, recordMap.Realm, realmName)
	testutils.AssertEquals(t, recordMap.Event, eventName)
	testutils.AssertEquals(t, recordMap.SplitEventField, "contracts.$.data.products")
	testutils.AssertEquals(t, len(recordMap.Id), 2)
	testutils.AssertEquals(t, recordMap.Id[0].Unique, true)
	testutils.AssertEquals(t, recordMap.Id[1].Unique, true)
}

func TestGetRecordMapForLineRecordSuccess(t *testing.T) {
	eventName := "active-coupons"
	realmName := "bonuz-br"
	eventData, eventLocation, err := testutils.GetEventDataFromJsonFiles(eventName, realmName)
	if err != nil {
		t.Error(err)
		return
	}
	dbMock := &testutils.DBMock{
		LocationToReturn:         eventLocation,
		EventMappingDataToReturn: eventData,
	}
	recordMap, eventConfig, err := GetRecordMap(dbMock, realmName, eventName)
	if err != nil {
		t.Error(err)
		return
	}
	testutils.AssertEquals(t, recordMap.IdIndex[0].Index, 0)
	testutils.AssertEquals(t, recordMap.Location.String(), eventLocation)
	testutils.AssertEquals(t, recordMap.Date.Index, 1)
	testutils.AssertEquals(t, len(recordMap.Customer), 2)
	testutils.AssertEquals(t, len(recordMap.Modifiers), 3)
	testutils.AssertEquals(t, len(recordMap.GroupBy), 1)
	testutils.AssertEquals(t, len(recordMap.BillingScope), 2)
	testutils.AssertEquals(t, len(recordMap.Service), 1)
	testutils.AssertEquals(t, len(recordMap.Info), 2)
	testutils.AssertEquals(t, len(recordMap.Metrics), 2)
	testutils.AssertEquals(t, recordMap.Customer[0].Index, 2)
	testutils.AssertEquals(t, recordMap.Customer[0].Name, "accountCode")
	testutils.AssertEquals(t, recordMap.Customer[1].Index, 3)
	testutils.AssertEquals(t, recordMap.Customer[1].Name, "experience")
	testutils.AssertEquals(t, recordMap.Modifiers[0].Index, 3)
	testutils.AssertEquals(t, recordMap.Modifiers[0].Name, "experience")
	testutils.AssertEquals(t, recordMap.GroupBy[0].Index, 3)
	testutils.AssertEquals(t, recordMap.GroupBy[0].Name, "experience")
	testutils.AssertEquals(t, recordMap.BillingScope[0].Index, 3)
	testutils.AssertEquals(t, recordMap.BillingScope[0].Name, "experience")
	testutils.AssertEquals(t, recordMap.Modifiers[1].Index, 4)
	testutils.AssertEquals(t, recordMap.Modifiers[1].Name, "alliance")
	testutils.AssertEquals(t, recordMap.Service[0].Index, 5)
	testutils.AssertEquals(t, recordMap.Service[0].Name, "prize")
	testutils.AssertEquals(t, recordMap.Modifiers[2].Index, 5)
	testutils.AssertEquals(t, recordMap.Modifiers[2].Name, "prize")
	testutils.AssertEquals(t, recordMap.BillingScope[1].Index, 5)
	testutils.AssertEquals(t, recordMap.BillingScope[1].Name, "prize")
	testutils.AssertEquals(t, recordMap.Info[0].Index, 6)
	testutils.AssertEquals(t, recordMap.Info[0].Name, "regionCode")
	testutils.AssertEquals(t, recordMap.Info[1].Index, 7)
	testutils.AssertEquals(t, recordMap.Info[1].Name, "portedFrom")
	testutils.AssertEquals(t, recordMap.Metrics[0].Index, 8)
	testutils.AssertEquals(t, recordMap.Metrics[0].Name, "couponValue")
	testutils.AssertEquals(t, recordMap.Metrics[1].Index, 9)
	testutils.AssertEquals(t, recordMap.Metrics[1].Name, "couponAmount")
	testutils.AssertEquals(t, recordMap.CountIndex, 9)
	testutils.AssertEquals(t, recordMap.TpName, "")
	testutils.AssertEquals(t, fmt.Sprintf("%+v", eventConfig), fmt.Sprintf("%+v", eventData))
}

// Deprecated
func TestGetRecordMapWithTpNameSuccess(t *testing.T) {
	eventName := "alliance-active-coupons"
	realmName := "bonuz-br"
	eventData, eventLocation, err := testutils.GetEventDataFromJsonFiles(eventName, realmName)
	dbMock := &testutils.DBMock{
		LocationToReturn:         eventLocation,
		EventMappingDataToReturn: eventData,
	}
	recordMap, _, err := GetRecordMap(dbMock, realmName, eventName)
	if err != nil {
		t.Error(err)
		return
	}
	testutils.AssertEquals(t, recordMap.TpName, "active-coupons")
	testutils.AssertEquals(t, recordMap.Realm, "bonuz-br")
	testutils.AssertEquals(t, recordMap.Event, "alliance-active-coupons")
}

func TestGetRecordMapWithFilterSuccess(t *testing.T) {
	eventName := "filter-test"
	realmName := "bonuz-br"
	eventData, eventLocation, err := testutils.GetEventDataFromJsonFiles(eventName, realmName)
	if err != nil {
		t.Error(err)
		return
	}
	dbMock := &testutils.DBMock{
		LocationToReturn:         eventLocation,
		EventMappingDataToReturn: eventData,
	}
	recordMap, _, err := GetRecordMap(dbMock, realmName, eventName)
	if err != nil {
		t.Error(err)
		return
	}
	testutils.AssertEquals(t, len(recordMap.FieldFilters), 2)
	testutils.AssertEquals(t, recordMap.FieldFilters[0].Index, 3)
	testutils.AssertEquals(t, recordMap.FieldFilters[0].Criteria, "not-equals")
	testutils.AssertEquals(t, len(recordMap.FieldFilters[0].Values), 2)
	testutils.AssertEquals(t, recordMap.FieldFilters[0].Values["42.660.634/0001-10"], true)
	testutils.AssertEquals(t, recordMap.FieldFilters[0].Values["22.003.653/0001-67"], true)
	testutils.AssertEquals(t, recordMap.FieldFilters[1].Index, 5)
	testutils.AssertEquals(t, recordMap.FieldFilters[1].Criteria, "equals")
	testutils.AssertEquals(t, len(recordMap.FieldFilters[1].Values), 1)
	testutils.AssertEquals(t, recordMap.FieldFilters[1].Values["offer-transaction"], true)
}

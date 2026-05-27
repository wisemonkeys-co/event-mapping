package utils

import (
	"fmt"
	"strings"
	"testing"
	"time"

	testutils "github.com/wisemonkeys-co/event-mapping/test-utils"
	"github.com/wisemonkeys-co/event-mapping/types"
)

var loc, _ = time.LoadLocation("America/Sao_Paulo")
var recorMap = types.RecordMap{
	Realm:    "minu-shop-br",
	Event:    "event-streaming-minu-offer-order-transaction",
	Location: loc,
	Id: []types.RecordField{
		{Index: 0,
			Name:       "id",
			RemoteName: "order.id",
			Format:     "string",
			Value:      "",
		},
	},
	Date: types.RecordField{
		Index:      1,
		Name:       "date",
		RemoteName: "date",
		Format:     "2006-01-02T15:04:05.000Z",
		Value:      "",
	},
	Count: types.RecordField{
		Index:      0,
		Name:       "",
		RemoteName: "",
		Format:     "",
		Value:      "",
	},
	TpName: "",
	Customer: []types.RecordField{
		{
			Index:      2,
			Name:       "companyId",
			RemoteName: "company.id",
			Format:     "string",
			Value:      "",
		},
		{
			Index:      4,
			Name:       "accountId",
			RemoteName: "accountId",
			Format:     "string",
			Value:      "",
		},
	},
	TPCustomer: []types.RecordField{},
	Modifiers: []types.RecordField{
		{
			Index:      7,
			Name:       "paymentMethod",
			RemoteName: "order.paymentMethod",
			Format:     "string",
			Value:      "",
		},
	},
	Info: []types.RecordField{
		{
			Index:      3,
			Name:       "companyCNPJ",
			RemoteName: "company.cnpj",
			Format:     "string",
			Value:      "",
		},
	},
	Service: []types.RecordField{
		{
			Index:      5,
			Name:       "serviceId",
			RemoteName: "serviceId",
			Format:     "string",
			Value:      "",
		},
	},
	Metrics: []types.RecordField{
		{
			Index:      8,
			Name:       "orderAmount",
			RemoteName: "order.value",
			Format:     "string",
			Value:      "",
		},
	},
	Values: []types.RecordField{
		{
			Index:      8,
			Name:       "orderAmount",
			RemoteName: "order.value",
			Format:     "string",
			Value:      "",
		},
	},
	GroupBy: []types.RecordField{
		{
			Index:      6,
			Name:       "program",
			RemoteName: "program",
			Format:     "string",
			Value:      "",
		},
	},
	BillingScope: []types.RecordField{
		{
			Index:      6,
			Name:       "program",
			RemoteName: "program",
			Format:     "string",
			Value:      "",
		},
	},
}

func TestBuildRealmEventSuccess(t *testing.T) {
	files, err := testutils.ReadJsonFiles()
	if err != nil {
		t.Error(err)
		return
	}
	eventData := files["external-events"][0].(map[string]any)
	realmData := files["realms"][0].(types.Realm)
	realmEvent, err := BuildRealmEventFromMap(eventData, realmData.Events[0], recorMap)
	if err != nil {
		t.Error(err)
		return
	}
	testutils.AssertEquals(t, realmEvent.Realm, recorMap.Realm)
	testutils.AssertEquals(t, realmEvent.RealmEvent, recorMap.Event)
	testutils.AssertEquals(t, realmEvent.RealmID, "654d11c40aaa37fada8f7436")
	testutils.AssertEquals(t, realmEvent.RealmDate.Format(recorMap.Date.Format), "2023-10-02T09:00:00.000Z")
	testutils.AssertEquals(t, len(realmEvent.Customer), 2)
	testutils.AssertEquals(t, realmEvent.Customer[0], "63167b308e1ef661f7b7abcd")
	testutils.AssertEquals(t, realmEvent.Customer[1], "63167b308e1ef661f7b7e913")
	testutils.AssertEquals(t, len(realmEvent.Metric), 2)
	testutils.AssertEquals(t, realmEvent.Metric[0]["orderAmount"], float64(100))
	testutils.AssertEquals(t, realmEvent.Metric[1]["count"], float64(1))
	testutils.AssertEquals(t, len(realmEvent.GroupBy), 1)
	testutils.AssertEquals(t, realmEvent.GroupBy[0], "minu-offer")
	testutils.AssertEquals(t, len(realmEvent.BillingScope), 1)
	testutils.AssertEquals(t, realmEvent.BillingScope[0], "minu-offer")
	testutils.AssertEquals(t, len(realmEvent.Info), 1)
	testutils.AssertEquals(t, realmEvent.Info[0]["companyCNPJ"], "000.000.0001-00")
}

func TestBuildRealmEventShouldNotDrop(t *testing.T) {
	eventName := "event-streaming-minu-offer-order-transaction"
	realmName := "minu-shop-br"
	files, err := testutils.ReadJsonFiles()
	if err != nil {
		t.Error(err)
		return
	}
	eventData, eventLocation, err := testutils.GetEventDataFromJsonFiles(eventName, realmName)
	if err != nil {
		t.Error(err)
		return
	}
	eventMap := files["external-events"][0].(map[string]any)
	eventMap["company"] = map[string]interface{}{
		"id":          "63167b308e1ef661f7b7abcd",
		"cnpj":        "08.939.312/0001-77",
		"companyName": "Wise Monkeys ltda",
		"fantasyName": "Wise Monkeys",
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
	realmEvent, errorData := BuildRealmEventFromMap(eventMap, eventConfig, recordMap)
	if errorData != nil {
		t.Error(errorData)
		return
	}
	if ShouldDropEventMapBased(eventMap, recordMap) {
		t.Errorf("should not drop the event")
		return
	}
	testutils.AssertEquals(t, realmEvent.Raw, "654d11c40aaa37fada8f7436;2023-10-02T09:00:00.000Z;63167b308e1ef661f7b7abcd;08.939.312/0001-77;63167b308e1ef661f7b7e913;minu-offer-transaction;minu-offer;credit_card;100")
	testutils.AssertEquals(t, realmEvent.Info[0]["companyCNPJ"], "08.939.312/0001-77")
}

func TestDroppedEventBasedOnMap(t *testing.T) {
	eventName := "event-streaming-minu-offer-order-transaction"
	realmName := "minu-shop-br"
	eventData, eventLocation, err := testutils.GetEventDataFromJsonFiles(eventName, realmName)
	if err != nil {
		t.Error(err)
		return
	}
	dbMock := &testutils.DBMock{
		LocationToReturn:         eventLocation,
		EventMappingDataToReturn: eventData,
	}
	eventMap := make(map[string]interface{})
	eventMap["company"] = map[string]interface{}{
		"id":          "63167b308e1ef661f7b7abcd",
		"cnpj":        "48.811.335/0001-16", // must be dropped
		"companyName": "Wise Monkeys ltda",
		"fantasyName": "Wise Monkeys",
	}
	recordMap, _, err := GetRecordMap(dbMock, realmName, eventName)
	if err != nil {
		t.Error(err)
		return
	}
	if !ShouldDropEventMapBased(eventMap, recordMap) {
		t.Errorf("event should be droped based on company.cnpj")
	}
}

func TestBuildEventWithNullField(t *testing.T) {
	eventName := "event-streaming-minu-offer-order-transaction"
	realmName := "minu-shop-br"
	files, err := testutils.ReadJsonFiles()
	if err != nil {
		t.Error(err)
		return
	}
	eventData, eventLocation, err := testutils.GetEventDataFromJsonFiles(eventName, realmName)
	if err != nil {
		t.Error(err)
		return
	}
	dbMock := &testutils.DBMock{
		LocationToReturn:         eventLocation,
		EventMappingDataToReturn: eventData,
	}
	eventMap := files["external-events"][0].(map[string]any)
	// missing field "company.cnpj" (type "info")
	eventMap["company"] = map[string]interface{}{
		"id":          "63167b308e1ef661f7b7abcd",
		"companyName": "Wise Monkeys ltda",
		"fantasyName": "Wise Monkeys",
	}
	recordMap, eventConfig, err := GetRecordMap(dbMock, realmName, eventName)
	if err != nil {
		t.Error(err)
		return
	}
	realmEvent, errorData := BuildRealmEventFromMap(eventMap, eventConfig, recordMap)
	if errorData != nil {
		t.Error(errorData)
		return
	}
	testutils.AssertEquals(t, realmEvent.Raw, "654d11c40aaa37fada8f7436;2023-10-02T09:00:00.000Z;63167b308e1ef661f7b7abcd;;63167b308e1ef661f7b7e913;minu-offer-transaction;minu-offer;credit_card;100")
}

func TestBuildRealmEventCalcFieldEventBasedOnMap(t *testing.T) {
	eventName := "test-fields-calc-with-expression-map-based"
	realmName := "bonuz-br"
	files, err := testutils.ReadJsonFiles()
	if err != nil {
		t.Error(err)
		return
	}
	eventData, eventLocation, err := testutils.GetEventDataFromJsonFiles(eventName, realmName)
	if err != nil {
		t.Error(err)
		return
	}
	eventMap := files["external-events"][1].(map[string]any)
	dbMock := &testutils.DBMock{
		LocationToReturn:         eventLocation,
		EventMappingDataToReturn: eventData,
	}
	recordMap, eventConfig, err := GetRecordMap(dbMock, realmName, eventName)
	if err != nil {
		t.Error(err)
		return
	}
	realmEvent, errorData := BuildRealmEventFromMap(eventMap, eventConfig, recordMap)
	if errorData != nil {
		t.Error(errorData)
		return
	}
	testutils.AssertEquals(t, realmEvent.Realm, "bonuz-br")
	testutils.AssertEquals(t, realmEvent.RealmEvent, "test-fields-calc-with-expression-map-based")
	testutils.AssertEquals(t, realmEvent.Variables["accountCode"], "abdf17")
	testutils.AssertEquals(t, realmEvent.Variables["multiplier"], int64(2))
	testutils.AssertEquals(t, realmEvent.Variables["serviceValue"], 159.97)
	testutils.AssertEquals(t, realmEvent.Value[2]["multiplier"], float64(2))
	testutils.AssertEquals(t, realmEvent.Raw, "123987465;2026-05-02T19:00:00.000;ACME LTDA;service-foo;abdf17;159.970000;100.050000")
	fmt.Println(realmEvent.Raw)
}

func TestBuildRealmEventCalcFieldEventBasedOnLineRecord(t *testing.T) {
	eventName := "test-fields-calc-with-expression-line-based"
	realmName := "bonuz-br"
	eventData, eventLocation, err := testutils.GetEventDataFromJsonFiles(eventName, realmName)
	if err != nil {
		t.Error(err)
		return
	}
	lineRecord := "123987465;2026-05-02T19:00:00.000;ACME LTDA;service-foo;abdf17;159.97"
	dbMock := &testutils.DBMock{
		LocationToReturn:         eventLocation,
		EventMappingDataToReturn: eventData,
	}
	recordMap, eventConfig, err := GetRecordMap(dbMock, realmName, eventName)
	if err != nil {
		t.Error(err)
		return
	}
	realmEvent, errorData := BuildRealmEventFromLineRecord(strings.Split(lineRecord, ";"), eventConfig, recordMap)
	if errorData != nil {
		t.Error(errorData)
		return
	}
	testutils.AssertEquals(t, realmEvent.Realm, "bonuz-br")
	testutils.AssertEquals(t, realmEvent.RealmEvent, "test-fields-calc-with-expression-line-based")
	testutils.AssertEquals(t, realmEvent.Variables["accountCode"], "abdf17")
	testutils.AssertEquals(t, realmEvent.Variables["multiplier"], 0.5)
	testutils.AssertEquals(t, realmEvent.Variables["serviceValue"], 159.97)
	testutils.AssertEquals(t, realmEvent.Value[2]["multiplier"], 0.5)
}

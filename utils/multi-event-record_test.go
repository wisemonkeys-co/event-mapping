package utils

import (
	"encoding/json"
	"errors"
	"testing"

	testutils "github.com/wisemonkeys-co/event-mapping/test-utils"
	"github.com/wisemonkeys-co/event-mapping/types"
)

// Helpers

func splitTestSetup(t *testing.T, eventName string) (types.RecordMap, types.Event) {
	t.Helper()
	realmName := "bonuz-br"
	eventData, eventLocation, err := testutils.GetEventDataFromJsonFiles(eventName, realmName)
	if err != nil {
		t.Fatal(err)
	}
	dbMock := &testutils.DBMock{
		LocationToReturn:         eventLocation,
		EventMappingDataToReturn: eventData,
	}
	recordMap, eventConfig, err := GetRecordMap(dbMock, realmName, eventName)
	if err != nil {
		t.Fatal(err)
	}
	return recordMap, eventConfig
}

func parseSplitTestEvent(t *testing.T, eventJsonStr string) map[string]any {
	t.Helper()
	var event map[string]any
	if err := json.Unmarshal([]byte(eventJsonStr), &event); err != nil {
		t.Fatal(err)
	}
	return event
}

func assertSplitResultLen(t *testing.T, events []map[string]any, errs []error, expected int) {
	t.Helper()
	testutils.AssertEquals(t, len(events), expected)
	testutils.AssertEquals(t, len(errs), expected)
}

// buildSplitTestRealmEvents asserts there is no split error and builds the RealmEvents (as the services do).
func buildSplitTestRealmEvents(t *testing.T, events []map[string]any, errs []error, eventConfig types.Event, recordMap types.RecordMap) []types.RealmEvent {
	t.Helper()
	testutils.AssertEquals(t, len(events), len(errs))
	realmEvents := make([]types.RealmEvent, 0, len(events))
	for i := range events {
		if errs[i] != nil {
			t.Fatalf("unexpected error on item %d: %s", i, errs[i].Error())
		}
		if events[i] == nil {
			t.Fatalf("nil event on item %d", i)
		}
		realmEvent, err := BuildRealmEventFromMap(events[i], eventConfig, recordMap)
		if err != nil {
			t.Fatalf("could not build the realm event %d: %s", i, err.Error())
		}
		realmEvents = append(realmEvents, realmEvent)
	}
	return realmEvents
}

func assertSplitError(t *testing.T, err error, target error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error %v, got nil", target)
	}
	if !errors.Is(err, target) {
		t.Fatalf("expected error %v, got %v", target, err)
	}
}

// Success scenarios (same scenarios of ocs-event-streaming and event-api)

func TestSplitMultiEventRecordArrayPath(t *testing.T) {
	recordMap, eventConfig := splitTestSetup(t, "test-split-array")
	event := parseSplitTestEvent(t, `{
		"_id": "123987465",
		"metadata": {
			"created": "2023-10-02T09:00:00.000",
			"ignored": 123
		},
		"client": {
			"name": "ACME LTDA"
		},
		"accountCode": "abdf17",
		"contracts": [{
			"value": 159.97,
			"data": {
				"product": "service-foo"
			}
		}, {
			"value": 60.28,
			"data": {
				"product": "service-bar"
			}
		}]
	}`)
	events, errs := SplitMultiEventRecord(event, recordMap)
	assertSplitResultLen(t, events, errs, 2)
	// The array is replaced by the item and the id receives the sufix
	testutils.AssertEquals(t, events[0]["_id"], "123987465_0")
	testutils.AssertEquals(t, events[1]["_id"], "123987465_1")
	testutils.AssertEquals(t, events[0]["contracts"].(map[string]any)["value"], 159.97)
	testutils.AssertEquals(t, events[1]["contracts"].(map[string]any)["value"], 60.28)
	realmEvents := buildSplitTestRealmEvents(t, events, errs, eventConfig, recordMap)
	testutils.AssertEquals(t, realmEvents[0].RealmID, "123987465_0")
	testutils.AssertEquals(t, realmEvents[1].RealmID, "123987465_1")
	testutils.AssertEquals(t, realmEvents[0].Customer[0], "ACME LTDA")
	testutils.AssertEquals(t, realmEvents[1].Customer[0], "ACME LTDA")
	testutils.AssertEquals(t, realmEvents[0].Metric[0]["serviceValue"], 159.97)
	testutils.AssertEquals(t, realmEvents[1].Metric[0]["serviceValue"], 60.28)
	testutils.AssertEquals(t, realmEvents[0].Service[0], "service-foo")
	testutils.AssertEquals(t, realmEvents[1].Service[0], "service-bar")
	testutils.AssertEquals(t, realmEvents[0].Raw, "123987465_0;2023-10-02T09:00:00.000;ACME LTDA;service-foo;abdf17;159.970000")
	testutils.AssertEquals(t, realmEvents[1].Raw, "123987465_1;2023-10-02T09:00:00.000;ACME LTDA;service-bar;abdf17;60.280000")
}

func TestSplitMultiEventRecordArrayPathMultiIds(t *testing.T) {
	recordMap, eventConfig := splitTestSetup(t, "test-split-array-multi-index")
	event := parseSplitTestEvent(t, `{
		"_id": "123987465",
		"metadata": {
			"created": "2023-10-02T09:00:00.000",
			"ignored": 123
		},
		"client": {
			"name": "ACME LTDA"
		},
		"accountCode": "abdf17",
		"contracts": [{
			"value": 159.97,
			"data": {
				"product": "service-foo"
			},
			"sufixId": "903"
		}, {
			"value": 60.28,
			"data": {
				"product": "service-bar"
			},
			"sufixId": 942
		}]
	}`)
	events, errs := SplitMultiEventRecord(event, recordMap)
	assertSplitResultLen(t, events, errs, 2)
	realmEvents := buildSplitTestRealmEvents(t, events, errs, eventConfig, recordMap)
	// An id inside the split path has value: no sufix
	testutils.AssertEquals(t, realmEvents[0].RealmID, "123987465|903")
	testutils.AssertEquals(t, realmEvents[1].RealmID, "123987465|942")
	testutils.AssertEquals(t, realmEvents[0].Raw, "123987465;2023-10-02T09:00:00.000;ACME LTDA;service-foo;abdf17;159.970000;903")
	testutils.AssertEquals(t, realmEvents[1].Raw, "123987465;2023-10-02T09:00:00.000;ACME LTDA;service-bar;abdf17;60.280000;942")
}

func TestSplitMultiEventRecordPopulateValuesProperties(t *testing.T) {
	recordMap, eventConfig := splitTestSetup(t, "test-value-variables")
	event := parseSplitTestEvent(t, `{
		"_id": "123987465",
		"metadata": {
			"created": "2023-10-02T09:00:00.000",
			"ignored": 123
		},
		"client": {
			"name": "ACME LTDA"
		},
		"accountCode": "abdf17",
		"contracts": [{
			"value": 159.97,
			"data": {
				"product": "service-foo"
			},
			"intValue": 903,
			"flag": true
		}, {
			"value": 60.28,
			"data": {
				"product": "service-bar"
			},
			"intValue": 942,
			"flag": false
		}]
	}`)
	events, errs := SplitMultiEventRecord(event, recordMap)
	assertSplitResultLen(t, events, errs, 2)
	realmEvents := buildSplitTestRealmEvents(t, events, errs, eventConfig, recordMap)
	testutils.AssertEquals(t, realmEvents[0].Variables["valueFlag"], true)
	testutils.AssertEquals(t, realmEvents[0].Variables["intValue"], int64(903))
	testutils.AssertEquals(t, realmEvents[0].Variables["serviceValue"], 159.97)
	testutils.AssertEquals(t, realmEvents[1].Variables["valueFlag"], false)
	testutils.AssertEquals(t, realmEvents[1].Variables["intValue"], int64(942))
	testutils.AssertEquals(t, realmEvents[1].Variables["serviceValue"], 60.28)
}

func TestSplitMultiEventRecordNestedArray(t *testing.T) {
	recordMap, eventConfig := splitTestSetup(t, "test-split-neasted-array-with-variables")
	event := parseSplitTestEvent(t, `{
		"_id": "123987465",
		"metadata": {
			"created": "2023-10-02T09:00:00.000",
			"ignored": 123
		},
		"client": {
			"name": "ACME LTDA"
		},
		"accountCode": "abdf17",
		"contracts": [{
			"registration": "987qwe",
			"value": 159.97,
			"data": {
				"products": [{
					"sku": "abc-xyz"
				}, {
					"sku": "123-456"
				}]
			},
			"intValue": 903,
			"flag": true
		}, {
			"registration": "poi654",
			"value": 60.28,
			"data": {
				"products": [{
					"sku": "def-klm"
				}, {
					"sku": "369-852"
				}]
			},
			"intValue": 942,
			"flag": false
		}]
	}`)
	events, errs := SplitMultiEventRecord(event, recordMap)
	assertSplitResultLen(t, events, errs, 4)
	// Both array levels are replaced by the selected items
	product := events[3]["contracts"].(map[string]any)["data"].(map[string]any)["products"].(map[string]any)
	testutils.AssertEquals(t, product["sku"], "369-852")
	realmEvents := buildSplitTestRealmEvents(t, events, errs, eventConfig, recordMap)
	for _, realmEvent := range realmEvents {
		testutils.AssertEquals(t, len(realmEvent.Service), 1)
	}
	testutils.AssertEquals(t, realmEvents[0].RealmID, "123987465|987qwe_0_0")
	testutils.AssertEquals(t, realmEvents[1].RealmID, "123987465|987qwe_0_1")
	testutils.AssertEquals(t, realmEvents[2].RealmID, "123987465|poi654_1_0")
	testutils.AssertEquals(t, realmEvents[3].RealmID, "123987465|poi654_1_1")
	testutils.AssertEquals(t, realmEvents[0].Service[0], "abc-xyz")
	testutils.AssertEquals(t, realmEvents[1].Service[0], "123-456")
	testutils.AssertEquals(t, realmEvents[2].Service[0], "def-klm")
	testutils.AssertEquals(t, realmEvents[3].Service[0], "369-852")
	testutils.AssertEquals(t, realmEvents[0].Variables["intValue"], int64(903))
	testutils.AssertEquals(t, realmEvents[3].Variables["intValue"], int64(942))
	testutils.AssertEquals(t, realmEvents[0].Info[0]["intValue"], "903")
	testutils.AssertEquals(t, realmEvents[3].Info[0]["intValue"], "942")
	testutils.AssertEquals(t, realmEvents[0].Raw, "123987465;2023-10-02T09:00:00.000;ACME LTDA;abc-xyz;abdf17;159.970000;903;987qwe_0_0")
	testutils.AssertEquals(t, realmEvents[1].Raw, "123987465;2023-10-02T09:00:00.000;ACME LTDA;123-456;abdf17;159.970000;903;987qwe_0_1")
	testutils.AssertEquals(t, realmEvents[2].Raw, "123987465;2023-10-02T09:00:00.000;ACME LTDA;def-klm;abdf17;60.280000;942;poi654_1_0")
	testutils.AssertEquals(t, realmEvents[3].Raw, "123987465;2023-10-02T09:00:00.000;ACME LTDA;369-852;abdf17;60.280000;942;poi654_1_1")
}

func TestSplitMultiEventRecordNestedArrayWithNestedId(t *testing.T) {
	recordMap, eventConfig := splitTestSetup(t, "test-split-neasted-array-with-neasted-id")
	event := parseSplitTestEvent(t, `{
		"_id": "123987465",
		"metadata": {
			"created": "2023-10-02T09:00:00.000"
		},
		"client": {
			"name": "ACME LTDA"
		},
		"accountCode": "abdf17",
		"contracts": [{
			"value": 159.97,
			"data": {
				"products": [{
					"sku": "cba-321",
					"service": "foo"
				}, {
					"sku": "987-zyx",
					"service": "oof"
				}]
			}
		}, {
			"value": 60.28,
			"data": {
				"products": [{
					"sku": "654-qwe",
					"service": "bar"
				}, {
					"sku": "oiu-!@#",
					"service": "rab"
				}]
			}
		}]
	}`)
	events, errs := SplitMultiEventRecord(event, recordMap)
	assertSplitResultLen(t, events, errs, 4)
	realmEvents := buildSplitTestRealmEvents(t, events, errs, eventConfig, recordMap)
	testutils.AssertEquals(t, realmEvents[0].RealmID, "cba-321|123987465")
	testutils.AssertEquals(t, realmEvents[1].RealmID, "987-zyx|123987465")
	testutils.AssertEquals(t, realmEvents[2].RealmID, "654-qwe|123987465")
	testutils.AssertEquals(t, realmEvents[3].RealmID, "oiu-!@#|123987465")
	testutils.AssertEquals(t, realmEvents[0].Raw, "cba-321;2023-10-02T09:00:00.000;ACME LTDA;123987465;abdf17;foo;159.970000")
	testutils.AssertEquals(t, realmEvents[3].Raw, "oiu-!@#;2023-10-02T09:00:00.000;ACME LTDA;123987465;abdf17;rab;60.280000")
}

func TestSplitMultiEventRecordMissingValues(t *testing.T) {
	recordMap, eventConfig := splitTestSetup(t, "test-split-array-multi-index")
	event := parseSplitTestEvent(t, `{
		"_id": "123987465",
		"metadata": {
			"created": "2023-10-02T09:00:00.000",
			"ignored": 123
		},
		"client": {},
		"accountCode": "abdf17",
		"contracts": [{
			"value": null,
			"data": null,
			"sufixId": null
		}]
	}`)
	events, errs := SplitMultiEventRecord(event, recordMap)
	assertSplitResultLen(t, events, errs, 1)
	realmEvents := buildSplitTestRealmEvents(t, events, errs, eventConfig, recordMap)
	testutils.AssertEquals(t, realmEvents[0].RealmID, "123987465|<nil>_0")
	testutils.AssertEquals(t, realmEvents[0].Raw, "123987465;2023-10-02T09:00:00.000;;;abdf17;;<nil>_0")
}

func TestSplitMultiEventRecordIdsOnlyOnRoot(t *testing.T) {
	recordMap, eventConfig := splitTestSetup(t, "test-id-on-root")
	event := parseSplitTestEvent(t, `{
		"_id": 999999999999,
		"metadata": {
			"created": "2023-10-02T09:00:00.000",
			"ignored": 123
		},
		"client": {
			"name": "ACME LTDA"
		},
		"accountCode": "abdf17",
		"contracts": [{
			"data": {
				"product": "service-foo"
			}
		}, {
			"data": {
				"product": "service-bar"
			}
		}]
	}`)
	events, errs := SplitMultiEventRecord(event, recordMap)
	assertSplitResultLen(t, events, errs, 2)
	realmEvents := buildSplitTestRealmEvents(t, events, errs, eventConfig, recordMap)
	testutils.AssertEquals(t, realmEvents[0].RealmID, "999999999999_0|ACME LTDA")
	testutils.AssertEquals(t, realmEvents[1].RealmID, "999999999999_1|ACME LTDA")
	testutils.AssertEquals(t, realmEvents[0].Raw, "999999999999_0;2023-10-02T09:00:00.000;ACME LTDA;service-foo;abdf17")
	testutils.AssertEquals(t, realmEvents[1].Raw, "999999999999_1;2023-10-02T09:00:00.000;ACME LTDA;service-bar;abdf17")
}

// A root id field whose name starts with the split field name ("contractsBatchId" x "contracts")
// is not inside the split path: the sufix must be appended.
func TestSplitMultiEventRecordIdPrefixOnRoot(t *testing.T) {
	recordMap, eventConfig := splitTestSetup(t, "test-split-array-id-prefix-on-root")
	event := parseSplitTestEvent(t, `{
		"contractsBatchId": "B-77",
		"metadata": {
			"created": "2023-10-02T09:00:00.000"
		},
		"contracts": [{
			"value": 159.97,
			"data": {
				"product": "service-foo"
			}
		}, {
			"value": 60.28,
			"data": {
				"product": "service-bar"
			}
		}]
	}`)
	events, errs := SplitMultiEventRecord(event, recordMap)
	assertSplitResultLen(t, events, errs, 2)
	realmEvents := buildSplitTestRealmEvents(t, events, errs, eventConfig, recordMap)
	testutils.AssertEquals(t, realmEvents[0].RealmID, "B-77_0")
	testutils.AssertEquals(t, realmEvents[1].RealmID, "B-77_1")
	testutils.AssertEquals(t, realmEvents[0].Raw, "B-77_0;2023-10-02T09:00:00.000;service-foo;159.970000")
	testutils.AssertEquals(t, realmEvents[1].Raw, "B-77_1;2023-10-02T09:00:00.000;service-bar;60.280000")
}

func TestIsInsidePath(t *testing.T) {
	testutils.AssertEquals(t, isInsidePath("contracts", "contracts"), true)
	testutils.AssertEquals(t, isInsidePath("contracts.id", "contracts"), true)
	testutils.AssertEquals(t, isInsidePath("contracts.data.products.sku", "contracts.data.products"), true)
	testutils.AssertEquals(t, isInsidePath("contractsBatchId", "contracts"), false)
	testutils.AssertEquals(t, isInsidePath("contracts.itemsTotal", "contracts.items"), false)
	testutils.AssertEquals(t, isInsidePath("_id", "contracts"), false)
}

// Copy guarantees

func TestSplitMultiEventRecordDoesNotChangeTheRecord(t *testing.T) {
	recordMap, _ := splitTestSetup(t, "test-split-array-multi-index")
	eventJsonStr := `{
		"_id": "123987465",
		"metadata": {
			"created": "2023-10-02T09:00:00.000"
		},
		"contracts": [{
			"value": 1,
			"sufixId": null
		}, {
			"value": 2,
			"sufixId": null
		}]
	}`
	event := parseSplitTestEvent(t, eventJsonStr)
	before, _ := json.Marshal(event)
	events, errs := SplitMultiEventRecord(event, recordMap)
	assertSplitResultLen(t, events, errs, 2)
	after, _ := json.Marshal(event)
	// The sufix was written inside the items of the copies, not in the received record
	testutils.AssertEquals(t, string(after), string(before))
	testutils.AssertEquals(t, events[0]["contracts"].(map[string]any)["sufixId"], "<nil>_0")
	testutils.AssertEquals(t, events[1]["contracts"].(map[string]any)["sufixId"], "<nil>_1")
	// The extracted events are independent
	events[0]["metadata"].(map[string]any)["created"] = "changed"
	testutils.AssertEquals(t, events[1]["metadata"].(map[string]any)["created"], "2023-10-02T09:00:00.000")
	testutils.AssertEquals(t, event["metadata"].(map[string]any)["created"], "2023-10-02T09:00:00.000")
}

func TestSplitMultiEventRecordNormalizesTypes(t *testing.T) {
	recordMap, _ := splitTestSetup(t, "test-split-array")
	// Typed slices (e.g. built by a decoder) are accepted
	event := map[string]any{
		"_id": "abc",
		"contracts": []map[string]any{
			{"value": 1, "data": map[string]any{"product": "foo"}},
			{"value": int64(2), "data": map[string]any{"product": "bar"}},
		},
	}
	events, errs := SplitMultiEventRecord(event, recordMap)
	assertSplitResultLen(t, events, errs, 2)
	testutils.AssertEquals(t, errs[0], nil)
	testutils.AssertEquals(t, errs[1], nil)
	testutils.AssertEquals(t, events[0]["_id"], "abc_0")
	testutils.AssertEquals(t, events[1]["contracts"].(map[string]any)["value"], 2.0)
}

// Error scenarios

func TestSplitMultiEventRecordInvalidSplitEventField(t *testing.T) {
	event := parseSplitTestEvent(t, `{"contracts": [{"value": 1}]}`)
	for _, splitEventField := range []string{"", ".$.contracts", "contracts.$."} {
		events, errs := SplitMultiEventRecord(event, types.RecordMap{SplitEventField: splitEventField})
		assertSplitResultLen(t, events, errs, 1)
		testutils.AssertEquals(t, events[0] == nil, true)
		assertSplitError(t, errs[0], ErrInvalidSplitEventField)
	}
}

func TestSplitMultiEventRecordArrayNotFound(t *testing.T) {
	recordMap, _ := splitTestSetup(t, "test-split-array")
	event := parseSplitTestEvent(t, `{"_id": "123987465", "items": [{"value": 1}]}`)
	events, errs := SplitMultiEventRecord(event, recordMap)
	assertSplitResultLen(t, events, errs, 1)
	testutils.AssertEquals(t, events[0] == nil, true)
	assertSplitError(t, errs[0], ErrSplitArrayNotFound)
}

func TestSplitMultiEventRecordNilRecord(t *testing.T) {
	recordMap, _ := splitTestSetup(t, "test-split-array")
	events, errs := SplitMultiEventRecord(nil, recordMap)
	assertSplitResultLen(t, events, errs, 1)
	assertSplitError(t, errs[0], ErrSplitArrayNotFound)
}

func TestSplitMultiEventRecordPathNotArray(t *testing.T) {
	recordMap, _ := splitTestSetup(t, "test-split-array")
	for _, eventJsonStr := range []string{
		`{"_id": "1", "contracts": {"value": 1}}`,
		`{"_id": "1", "contracts": "foo"}`,
		`{"_id": "1", "contracts": null}`,
	} {
		events, errs := SplitMultiEventRecord(parseSplitTestEvent(t, eventJsonStr), recordMap)
		assertSplitResultLen(t, events, errs, 1)
		testutils.AssertEquals(t, events[0] == nil, true)
		assertSplitError(t, errs[0], ErrSplitPathNotArray)
	}
}

func TestSplitMultiEventRecordEmptyArray(t *testing.T) {
	recordMap, _ := splitTestSetup(t, "test-split-array")
	events, errs := SplitMultiEventRecord(parseSplitTestEvent(t, `{"_id": "1", "contracts": []}`), recordMap)
	assertSplitResultLen(t, events, errs, 0)
}

func TestSplitMultiEventRecordNestedErrorsKeepTheOtherItems(t *testing.T) {
	recordMap, _ := splitTestSetup(t, "test-split-neasted-array-with-variables")
	event := parseSplitTestEvent(t, `{
		"_id": "123987465",
		"metadata": {
			"created": "2023-10-02T09:00:00.000"
		},
		"contracts": [{
			"registration": "r0",
			"data": {
				"products": [{"sku": "a"}, {"sku": "b"}]
			}
		}, {
			"registration": "r1",
			"data": {}
		},
		"not-an-object",
		{
			"registration": "r3",
			"data": {
				"products": "not-an-array"
			}
		}, {
			"registration": "r4",
			"data": {
				"products": [{"sku": "c"}]
			}
		}, {
			"registration": "r5",
			"data": {
				"products": []
			}
		}]
	}`)
	events, errs := SplitMultiEventRecord(event, recordMap)
	// 0_0, 0_1, item 1 (error), item 2 (error), item 3 (error), 4_0 (item 5 is empty)
	assertSplitResultLen(t, events, errs, 6)
	for _, i := range []int{0, 1, 5} {
		testutils.AssertEquals(t, errs[i], nil)
		if events[i] == nil {
			t.Fatalf("nil event on item %d", i)
		}
	}
	for _, i := range []int{2, 3, 4} {
		testutils.AssertEquals(t, events[i] == nil, true)
	}
	assertSplitError(t, errs[2], ErrSplitArrayNotFound)
	assertSplitError(t, errs[3], ErrSplitItemNotObject)
	assertSplitError(t, errs[4], ErrSplitPathNotArray)
	// The sufixes keep the original positions
	testutils.AssertEquals(t, events[0]["contracts"].(map[string]any)["registration"], "r0_0_0")
	testutils.AssertEquals(t, events[1]["contracts"].(map[string]any)["registration"], "r0_0_1")
	testutils.AssertEquals(t, events[5]["contracts"].(map[string]any)["registration"], "r4_4_0")
	testutils.AssertEquals(t, events[5]["contracts"].(map[string]any)["data"].(map[string]any)["products"].(map[string]any)["sku"], "c")
}

func TestSplitMultiEventRecordCopyError(t *testing.T) {
	recordMap, _ := splitTestSetup(t, "test-split-array")
	// Functions can not be copied (JSON)
	event := map[string]any{
		"_id":       "1",
		"contracts": []any{map[string]any{"value": 1}},
		"invalid":   func() {},
	}
	events, errs := SplitMultiEventRecord(event, recordMap)
	assertSplitResultLen(t, events, errs, 1)
	assertSplitError(t, errs[0], ErrSplitEventCopy)
}

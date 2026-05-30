package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/wisemonkeys-co/goval"

	"github.com/wisemonkeys-co/event-mapping/types"
)

var eval = goval.NewEvaluator()

// From map
func BuildRealmEventFromMap(event map[string]any, eventConfig types.Event, recordMap types.RecordMap) (realmEvent types.RealmEvent, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	realmEvent.Realm = recordMap.Realm
	realmEvent.RealmEvent = recordMap.Event
	realmEvent.RealmID = extractStringFromMap(event, recordMap.Id, "|")
	realmEvent.RealmDate, err = GetRealmEventDate(event, recordMap)
	if err != nil {
		return
	}
	if recordMap.TpName != "" {
		realmEvent.TpName = recordMap.TpName
	}
	count, counterFound := GetValueFromMap(recordMap.Count.RemoteName, event)
	if counterFound {
		realmEvent.Count = int32(count.(float64))
	} else {
		realmEvent.Count = 1
	}
	eventKVPair := mapEventMapToInternalObject(event, eventConfig)
	realmEvent.Customer = extractStringListFromMap(event, recordMap.Customer)
	realmEvent.TpCustomer = extractStringListFromMap(event, recordMap.TPCustomer)
	realmEvent.Metric = extractStringFloatMapListFromMap(event, recordMap.Metrics, eventKVPair)
	realmEvent.Metric = append(realmEvent.Metric, getFloatMap("count", realmEvent.Count))
	realmEvent.Value = extractStringFloatMapListFromMap(event, recordMap.Values, eventKVPair)
	realmEvent.Variables = extractInterfaceMapFromMap(event, recordMap.Variables, eventKVPair)
	realmEvent.Modifier = extractStringMapListFromMap(event, recordMap.Modifiers, eventKVPair)
	realmEvent.Service = extractStringListFromMap(event, recordMap.Service)
	realmEvent.GroupBy = extractStringListFromMap(event, recordMap.GroupBy)
	realmEvent.BillingScope = extractStringListFromMap(event, recordMap.BillingScope)
	realmEvent.Info = extractStringMapListFromMap(event, recordMap.Info, eventKVPair)
	realmEvent.Raw = GetEventLineStringFromMap(event, eventConfig)
	return
}

func GetValueFromMap(fieldName string, object map[string]any) (value any, found bool) {
	path := strings.SplitN(fieldName, ".", 2)
	if len(path) > 1 {
		if container, ok := object[path[0]].(map[string]any); ok {
			value, found = GetValueFromMap(path[1], container)
		}
	} else {
		value, found = object[path[0]]
	}

	return
}

func GetEventLineStringFromMap(event map[string]any, eventConfig types.Event) string {
	var values []string
	for _, c := range eventConfig.FieldMapping {
		if c.RemoteName != "" {
			value, _ := GetValueFromMap(c.RemoteName, event)
			values = append(values, GetString(value))
		}
	}
	return strings.Join(values, ";")
}

func ShouldDropEventMapBased(event map[string]any, recordMap types.RecordMap) bool {
	for i := 0; i < len(recordMap.FieldFilters); i++ {
		filter := recordMap.FieldFilters[i]
		value, ok := GetValueFromMap(filter.Path, event)
		if !ok {
			continue
		}
		hit := filter.Values[GetString(value)]
		if (filter.Criteria == "equals" && !hit) || (filter.Criteria == "not-equals" && hit) {
			return true
		}
	}
	return false
}

func mapEventMapToInternalObject(event map[string]any, eventConfig types.Event) map[string]interface{} {
	eventKVPair := make(map[string]interface{})
	for _, f := range eventConfig.FieldMapping {
		if f.RemoteName != "" {
			data, ok := GetValueFromMap(f.RemoteName, event)
			if ok {
				eventKVPair[f.Name] = data
			} else {
				eventKVPair[f.Name] = nil
			}
		}
	}
	return eventKVPair
}

func extractStringListFromMap(event map[string]any, mappedFields []types.RecordField) (strList []string) {
	strList = make([]string, 0)
	for _, mappedField := range mappedFields {
		value, ok := GetValueFromMap(mappedField.RemoteName, event)
		if ok {
			strList = append(strList, GetString(value))
		}
	}
	return
}

func extractStringFromMap(event map[string]any, mappedFields []types.RecordField, sep string) (str string) {
	strList := make([]string, 0)
	for _, mappedField := range mappedFields {
		value, ok := GetValueFromMap(mappedField.RemoteName, event)
		if ok {

			strList = append(strList, GetString(value))
		}
	}
	return strings.Join(strList, sep)
}

func extractStringMapListFromMap(event map[string]any, mappedFields []types.RecordField, eventKVPair map[string]interface{}) (strMapList []map[string]string) {
	strMapList = make([]map[string]string, 0)
	for _, mappedField := range mappedFields {
		var value any
		var ok bool
		if mappedField.Expression != "" {
			vars := make(map[string]any)
			vars["event"] = eventKVPair
			result, errEval := eval.Evaluate(mappedField.Expression, vars, MapFunctions())
			ok = errEval == nil
			value = result
		} else {
			value, ok = GetValueFromMap(mappedField.RemoteName, event)
		}
		if ok {
			strMap := make(map[string]string)
			strMap[mappedField.Name] = GetString(value)
			strMapList = append(strMapList, strMap)
		}
	}
	return
}

func extractStringFloatMapListFromMap(event map[string]any, mappedFields []types.RecordField, eventKVPair map[string]interface{}) (strFloatMapList []map[string]float64) {
	strFloatMapList = make([]map[string]float64, 0)
	for _, mappedField := range mappedFields {
		var value any
		var ok bool
		if mappedField.Expression != "" {
			vars := make(map[string]any)
			vars["event"] = eventKVPair
			result, errEval := eval.Evaluate(mappedField.Expression, vars, MapFunctions())
			ok = errEval == nil
			value = result
		} else {
			value, ok = GetValueFromMap(mappedField.RemoteName, event)
		}
		if ok && value != nil {
			floatMap := make(map[string]float64)
			floatMap[mappedField.Name], ok = value.(float64)
			if ok {
				strFloatMapList = append(strFloatMapList, floatMap)
			}
		}
	}
	return
}

func extractInterfaceMapFromMap(event map[string]any, fieldMap []types.RecordField, eventKVPair map[string]interface{}) map[string]any {
	m := make(map[string]any)
	for _, f := range fieldMap {
		var value any
		var ok bool
		if f.Expression != "" {
			vars := make(map[string]any)
			vars["event"] = eventKVPair
			result, errEval := eval.Evaluate(f.Expression, vars, MapFunctions())
			ok = errEval == nil
			value = result
		} else {
			value, ok = GetValueFromMap(f.RemoteName, event)
		}
		switch f.DataType {
		case "bool":
			if !ok {
				m[f.Name] = false
			}
			m[f.Name], ok = value.(bool)
			if !ok {
				m[f.Name] = false
			}
		case "float":
			if !ok {
				m[f.Name] = .0
			}
			m[f.Name], ok = value.(float64)
			if !ok {
				m[f.Name] = float64(0)
			}
		case "int":
			if !ok {
				m[f.Name] = int64(0)
			}
			num, ok := value.(float64)
			if !ok {
				m[f.Name] = int64(0)
			} else {
				m[f.Name] = int64(num)
			}
		case "string":
			if !ok {
				m[f.Name] = ""
			}
			m[f.Name], ok = value.(string)
			if !ok {
				m[f.Name] = ""
			}
		}
	}
	return m
}

// from line
func BuildRealmEventFromLineRecord(lineRecord []string, eventConfig types.Event, recordMap types.RecordMap) (realmEvent types.RealmEvent, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	realmEvent.Realm = recordMap.Realm
	realmEvent.RealmEvent = recordMap.Event
	realmEvent.RealmID = extractStringFromLineRecord(lineRecord, recordMap.Id, "|")
	d, err := time.ParseInLocation(recordMap.Date.Format, lineRecord[recordMap.Date.Index], recordMap.Location)
	if err != nil {
		return
	}
	realmEvent.RealmDate = d
	realmEvent.Count = 1
	// Melhorar este trecho . Campo do tipo count não é obrigatório (default = 1) . O valor padrão para int = 0 (recordMap.CountIndex). Remetendo a uma informação falsa
	if recordMap.CountIndex > 0 {
		i, errI := strconv.Atoi(lineRecord[recordMap.CountIndex])
		if errI != nil {
			return
		}
		realmEvent.Count = int32(i)
	}
	if recordMap.TpName != "" {
		realmEvent.TpName = recordMap.TpName
	}
	eventKVPair := mapLineRecordToInternalObject(lineRecord, eventConfig)
	realmEvent.Customer = extractStringListFromLineRecord(lineRecord, recordMap.Customer)
	realmEvent.TpCustomer = extractStringListFromLineRecord(lineRecord, recordMap.TPCustomer)
	realmEvent.Metric = extractStringFloatMapListFromLineRecord(lineRecord, recordMap.Metrics, eventKVPair)
	realmEvent.Metric = append(realmEvent.Metric, getFloatMap("count", realmEvent.Count))
	realmEvent.Value = extractStringFloatMapListFromLineRecord(lineRecord, recordMap.Values, eventKVPair)
	realmEvent.Variables = extractInterfaceMapFromLineRecord(lineRecord, recordMap.Variables, eventKVPair)
	realmEvent.Modifier = extractStringMapListFromLineRecord(lineRecord, recordMap.Modifiers, eventKVPair)
	realmEvent.Service = extractStringListFromLineRecord(lineRecord, recordMap.Service)
	realmEvent.GroupBy = extractStringListFromLineRecord(lineRecord, recordMap.GroupBy)
	realmEvent.BillingScope = extractStringListFromLineRecord(lineRecord, recordMap.BillingScope)
	realmEvent.Info = extractStringMapListFromLineRecord(lineRecord, recordMap.Info, eventKVPair)
	realmEvent.Raw = strings.Join(lineRecord, ";")
	return
}

func ShouldDropEventLineBased(lineRecord []string, recordMap types.RecordMap) bool {
	for i := 0; i < len(recordMap.FieldFilters); i++ {
		filter := recordMap.FieldFilters[i]
		hit := filter.Values[lineRecord[filter.Index]]
		if (filter.Criteria == "equals" && !hit) || (filter.Criteria == "not-equals" && hit) {
			return true
		}
	}
	return false
}

func mapLineRecordToInternalObject(lineRecord []string, eventConfig types.Event) map[string]interface{} {
	eventKVPair := make(map[string]interface{})
	for _, f := range eventConfig.FieldMapping {
		if f.RemoteName != "" {
			eventKVPair[f.Name] = lineRecord[f.Index]
		}
	}
	return eventKVPair
}

func extractInterfaceMapFromLineRecord(lineRecord []string, fieldMap []types.RecordField, eventKVPair map[string]interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	for _, f := range fieldMap {
		var err error
		var str string
		if f.Expression != "" {
			vars := make(map[string]any)
			vars["event"] = eventKVPair
			result, errEval := eval.Evaluate(f.Expression, vars, MapFunctions())
			if errEval == nil {
				str = fmt.Sprintf("%v", result)
			} else {
				err = errEval
			}
		} else {
			str = lineRecord[f.Index]
		}
		switch f.DataType {
		case "bool":
			m[f.Name], err = strconv.ParseBool(str)
			if err != nil {
				m[f.Name] = false
			}
		case "float":
			m[f.Name], err = strconv.ParseFloat(str, 64)
			if err != nil {
				m[f.Name] = 0
			}
		case "int":
			m[f.Name], err = strconv.ParseInt(str, 10, 64)
			if err != nil {
				m[f.Name] = 0
			}
		case "string":
			m[f.Name] = str
		}
	}
	return m
}

func extractStringListFromLineRecord(lineRecord []string, mappedFields []types.RecordField) (strList []string) {
	strList = make([]string, 0)
	for _, mappedField := range mappedFields {
		strList = append(strList, lineRecord[mappedField.Index])
	}
	return
}

func extractStringFromLineRecord(lineRecord []string, mappedFields []types.RecordField, sep string) (str string) {
	strList := make([]string, 0)
	for _, mappedField := range mappedFields {
		strList = append(strList, lineRecord[mappedField.Index])
	}
	return strings.Join(strList, sep)
}

func extractStringMapListFromLineRecord(lineRecord []string, mappedFields []types.RecordField, eventKVPair map[string]interface{}) (strMapList []map[string]string) {
	strMapList = make([]map[string]string, 0)
	for _, mappedField := range mappedFields {
		strMap := make(map[string]string)
		var str string
		if mappedField.Expression != "" {
			vars := make(map[string]any)
			vars["event"] = eventKVPair
			result, errEval := eval.Evaluate(mappedField.Expression, vars, MapFunctions())
			if errEval == nil {
				resultStr, ok := result.(string)
				if ok {
					str = resultStr
				} else {
					str = `<nil>`
				}
			} else {
				str = `<nil>`
			}
		} else {
			str = lineRecord[mappedField.Index]
		}
		strMap[mappedField.Name] = str
		strMapList = append(strMapList, strMap)
	}
	return
}

func extractStringFloatMapListFromLineRecord(lineRecord []string, mappedFields []types.RecordField, eventKVPair map[string]interface{}) (strFloatMapList []map[string]float64) {
	strFloatMapList = make([]map[string]float64, 0)
	for _, mappedField := range mappedFields {
		floatMap := make(map[string]float64)
		var str string
		if mappedField.Expression != "" {
			vars := make(map[string]any)
			vars["event"] = eventKVPair
			result, errEval := eval.Evaluate(mappedField.Expression, vars, MapFunctions())
			if errEval != nil {
				floatMap[mappedField.Name] = 0
			} else {
				floatMap[mappedField.Name] = result.(float64)
			}
		} else {
			str = lineRecord[mappedField.Index]
			fieldValue, err := strconv.ParseFloat(str, 64)
			// TODO Deve ser classificado como válido ou inválido?
			if err != nil {
				floatMap[mappedField.Name] = 0
			} else {
				floatMap[mappedField.Name] = fieldValue
			}
		}
		strFloatMapList = append(strFloatMapList, floatMap)
	}
	return
}

// Util
func GetRealmEventDate(event map[string]any, recordMap types.RecordMap) (date time.Time, err error) {
	realmDate, ok := GetValueFromMap(recordMap.Date.RemoteName, event)
	if !ok {
		err = fmt.Errorf("attribute %s not found", recordMap.Date.RemoteName)
		return
	}
	date, err = time.ParseInLocation(recordMap.Date.Format, realmDate.(string), recordMap.Location)
	if err != nil {
		err = fmt.Errorf("invalid Date on field %s", recordMap.Date.RemoteName)
	}
	return
}

func GetString(unk any) string {
	if unk == nil {
		return ""
	}
	switch i := unk.(type) {
	case float64:
		f := unk.(float64)
		if f == math.Trunc(f) {
			return strconv.Itoa(int(f))
		}
		return fmt.Sprintf("%f", i)
	case int:
		return strconv.Itoa(i)
	default:
		return fmt.Sprintf("%v", unk)
	}
}

func ValidateEventConfig(eventConfig types.Event) (bool, error) {
	firstExpressionIndex := -1
	for i, f := range eventConfig.FieldMapping {
		if f.RemoteName != "" {
			if firstExpressionIndex > -1 && i > firstExpressionIndex {
				return false, fmt.Errorf(
					"the expression item %d cannot have a index greater than the mapped field %d %s",
					firstExpressionIndex, f.Index, f.Name,
				)
			}
		} else {
			if f.Expression != "" {
				if firstExpressionIndex == -1 {
					firstExpressionIndex = i
				}
			} else {
				return false, fmt.Errorf("unknow field mapping for item %d %s", f.Index, f.Name)
			}
		}
	}
	return true, nil
}

// Common
func getFloatMap(name string, value int32) (floatMap map[string]float64) {
	floatMap = make(map[string]float64)
	floatMap[name] = float64(value)
	return
}

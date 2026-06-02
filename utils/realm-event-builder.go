package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/wisemonkeys-co/event-mapping/types"
)

// From map
func BuildRealmEventFromMap(event map[string]any, eventConfig types.Event, recordMap types.RecordMap) (realmEvent types.RealmEvent, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	realmEvent.Realm = recordMap.Realm
	realmEvent.RealmEvent = recordMap.Event
	realmEvent.RealmID = extractString(event, recordMap.Id, "|")
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
	realmEvent.Customer = extractStringListFromMap(event, recordMap.Customer)
	realmEvent.TpCustomer = extractStringListFromMap(event, recordMap.TPCustomer)
	realmEvent.Metric = extractStringFloatMapList(event, recordMap.Metrics)
	realmEvent.Metric = append(realmEvent.Metric, getFloatMap("count", realmEvent.Count))
	realmEvent.Value = extractStringFloatMapList(event, recordMap.Values)
	realmEvent.Variables = extractInterfaceMapFromMap(event, recordMap.Variables)
	realmEvent.Modifier = extractStringMapListFromMap(event, recordMap.Modifiers)
	realmEvent.Service = extractStringListFromMap(event, recordMap.Service)
	realmEvent.GroupBy = extractStringListFromMap(event, recordMap.GroupBy)
	realmEvent.BillingScope = extractStringListFromMap(event, recordMap.BillingScope)
	realmEvent.Info = extractStringMapListFromMap(event, recordMap.Info)
	realmEvent.Raw = GetEventLineString(event, eventConfig)
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

func GetEventLineString(event map[string]any, eventConfig types.Event) string {
	line := ""
	for i, c := range eventConfig.FieldMapping {
		if i > 0 {
			line += ";"
		}
		value, _ := GetValueFromMap(c.RemoteName, event)
		field := GetString(value)
		line += field
	}
	return line
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

func getFloatMap(name string, value int32) (floatMap map[string]float64) {
	floatMap = make(map[string]float64)
	floatMap[name] = float64(value)
	return
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

func extractString(event map[string]any, mappedFields []types.RecordField, sep string) (str string) {
	strList := make([]string, 0)
	for _, mappedField := range mappedFields {
		value, ok := GetValueFromMap(mappedField.RemoteName, event)
		if ok {

			strList = append(strList, GetString(value))
		}
	}
	return strings.Join(strList, sep)
}

func extractStringMapListFromMap(event map[string]any, mappedFields []types.RecordField) (strMapList []map[string]string) {
	strMapList = make([]map[string]string, 0)
	for _, mappedField := range mappedFields {
		value, ok := GetValueFromMap(mappedField.RemoteName, event)
		if ok {
			strMap := make(map[string]string)
			strMap[mappedField.Name] = GetString(value)
			strMapList = append(strMapList, strMap)
		}
	}
	return
}

func extractStringFloatMapList(event map[string]any, mappedFields []types.RecordField) (strFloatMapList []map[string]float64) {
	strFloatMapList = make([]map[string]float64, 0)
	for _, mappedField := range mappedFields {
		value, ok := GetValueFromMap(mappedField.RemoteName, event)
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

func extractInterfaceMapFromMap(event map[string]any, fieldMap []types.RecordField) map[string]any {
	m := make(map[string]any)
	for _, f := range fieldMap {
		var ok bool
		value, found := GetValueFromMap(f.RemoteName, event)
		switch f.DataType {
		case "bool":
			if !found {
				m[f.Name] = false
			}
			m[f.Name], ok = value.(bool)
			if !ok {
				m[f.Name] = false
			}
		case "float":
			if !found {
				m[f.Name] = .0
			}
			m[f.Name], ok = value.(float64)
			if !ok {
				m[f.Name] = float64(0)
			}
		case "int":
			if !found {
				m[f.Name] = int64(0)
			}
			num, ok := value.(float64)
			if !ok {
				m[f.Name] = int64(0)
			} else {
				m[f.Name] = int64(num)
			}
		case "string":
			if !found {
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

func BuildRealmEventFromLineRecord(lineRecord []string, recordMap types.RecordMap) (realmEvent types.RealmEvent, err error) {
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
	realmEvent.Customer = extractStringListFromLineRecord(lineRecord, recordMap.Customer)
	realmEvent.TpCustomer = extractStringListFromLineRecord(lineRecord, recordMap.TPCustomer)
	realmEvent.Metric = extractStringFloatMapListFromLineRecord(lineRecord, recordMap.Metrics)
	realmEvent.Metric = append(realmEvent.Metric, getFloatMap("count", realmEvent.Count))
	realmEvent.Value = extractStringFloatMapListFromLineRecord(lineRecord, recordMap.Values)
	realmEvent.Variables = extractInterfaceMapFromLineRecord(lineRecord, recordMap.Variables)
	realmEvent.Modifier = extractStringMapListFromLineRecord(lineRecord, recordMap.Modifiers)
	realmEvent.Service = extractStringListFromLineRecord(lineRecord, recordMap.Service)
	realmEvent.GroupBy = extractStringListFromLineRecord(lineRecord, recordMap.GroupBy)
	realmEvent.BillingScope = extractStringListFromLineRecord(lineRecord, recordMap.BillingScope)
	realmEvent.Info = extractStringMapListFromLineRecord(lineRecord, recordMap.Info)
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

func extractInterfaceMapFromLineRecord(lineRecord []string, fieldMap []types.RecordField) map[string]interface{} {
	m := make(map[string]interface{})
	for _, f := range fieldMap {
		var err error
		switch f.DataType {
		case "bool":
			m[f.Name], err = strconv.ParseBool(lineRecord[f.Index])
			if err != nil {
				m[f.Name] = false
			}
		case "float":
			m[f.Name], err = strconv.ParseFloat(lineRecord[f.Index], 64)
			if err != nil {
				m[f.Name] = 0
			}
		case "int":
			m[f.Name], err = strconv.ParseInt(lineRecord[f.Index], 10, 64)
			if err != nil {
				m[f.Name] = 0
			}
		case "string":
			m[f.Name] = lineRecord[f.Index]
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

func extractStringMapListFromLineRecord(lineRecord []string, mappedFields []types.RecordField) (strMapList []map[string]string) {
	strMapList = make([]map[string]string, 0)
	for _, mappedField := range mappedFields {
		strMap := make(map[string]string)
		strMap[mappedField.Name] = lineRecord[mappedField.Index]
		strMapList = append(strMapList, strMap)
	}
	return
}

func extractStringFloatMapListFromLineRecord(lineRecord []string, mappedFields []types.RecordField) (strFloatMapList []map[string]float64) {
	strFloatMapList = make([]map[string]float64, 0)
	for _, mappedField := range mappedFields {
		floatMap := make(map[string]float64)
		fieldValue, err := strconv.ParseFloat(lineRecord[mappedField.Index], 64)
		// TODO Deve ser classificado como válido ou inválido?
		if err != nil {
			floatMap[mappedField.Name] = 0
		} else {
			floatMap[mappedField.Name] = fieldValue
		}
		strFloatMapList = append(strFloatMapList, floatMap)
	}
	return
}

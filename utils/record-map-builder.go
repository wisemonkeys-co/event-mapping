package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/wisemonkeys-co/event-mapping/types"
)

func GetRecordMap(db DbInterface, realmName string, event string) (recMap types.RecordMap, eventConfig types.Event, err error) {
	eventConfig, location, err := db.FindRealmEventMappingData(realmName, event)
	if err != nil {
		return
	}
	recMap.Realm = realmName
	recMap.Event = event
	loc, _ := time.LoadLocation(location)
	recMap.Location = loc
	if eventConfig.TpName != "" {
		recMap.TpName = eventConfig.TpName
	}
	for i, f := range eventConfig.FieldMapping {
		if strings.Contains(f.RemoteName, ".$") {
			recMap.SplitEventField, err = selectNeastedArrayAddress(recMap.SplitEventField, f.RemoteName)
			if err != nil {
				err = fmt.Errorf("%s real name %s", err.Error(), realmName)
				return
			}
			eventConfig.FieldMapping[i].RemoteName = strings.Replace(f.RemoteName, ".$", "", -1)
			f.RemoteName = eventConfig.FieldMapping[i].RemoteName
		}
		unique := len(f.Types) == 1
		for _, t := range f.Types {
			switch t {
			case "id":
				recMap.Id = append(recMap.Id, types.RecordField{Index: f.Index, Name: f.Name, RemoteName: f.RemoteName, Format: f.Format, Unique: unique})
				recMap.IdIndex = append(recMap.IdIndex, types.RecordField{Index: f.Index, Name: f.Name, RemoteName: f.RemoteName, Format: f.Format})
			case "date":
				recMap.Date = types.RecordField{Index: f.Index, Name: f.Name, RemoteName: f.RemoteName, Format: f.Format}
			case "count":
				recMap.Count = types.RecordField{Index: f.Index, Name: f.Name, RemoteName: f.RemoteName, Format: f.Format}
				recMap.CountIndex = f.Index
			case "customer":
				recMap.Customer = append(recMap.Customer, types.RecordField{Index: f.Index, Name: f.Name, RemoteName: f.RemoteName, Format: f.Format})
			case "TPCustomer":
				recMap.TPCustomer = append(recMap.TPCustomer, types.RecordField{Index: f.Index, Name: f.Name, RemoteName: f.RemoteName, Format: f.Format})
			case "modifier":
				recMap.Modifiers = append(recMap.Modifiers, types.RecordField{Index: f.Index, Name: f.Name, RemoteName: f.RemoteName, Format: f.Format})
			case "service":
				recMap.Service = append(recMap.Service, types.RecordField{Index: f.Index, Name: f.Name, RemoteName: f.RemoteName, Format: f.Format})
			case "metric":
				recMap.Metrics = append(recMap.Metrics, types.RecordField{Index: f.Index, Name: f.Name, RemoteName: f.RemoteName, Format: f.Format})
				recMap.Variables = append(recMap.Variables, types.RecordField{Index: f.Index, Name: f.Name, RemoteName: f.RemoteName, Format: f.Format, Value: "", DataType: f.DataType})
			case "value":
				recMap.Values = append(recMap.Values, types.RecordField{Index: f.Index, Name: f.Name, RemoteName: f.RemoteName, Format: f.Format})
				recMap.Variables = append(recMap.Variables, types.RecordField{Index: f.Index, Name: f.Name, RemoteName: f.RemoteName, Format: f.Format, Value: "", DataType: f.DataType})
			case "info":
				recMap.Info = append(recMap.Info, types.RecordField{Index: f.Index, Name: f.Name, RemoteName: f.RemoteName, Format: f.Format})
			case "groupby":
				recMap.GroupBy = append(recMap.GroupBy, types.RecordField{Index: f.Index, Name: f.Name, RemoteName: f.RemoteName, Format: f.Format})
			case "billingScope":
				recMap.BillingScope = append(recMap.BillingScope, types.RecordField{Index: f.Index, Name: f.Name, RemoteName: f.RemoteName, Format: f.Format})
			}
		}
		if f.Filter != nil && f.Filter.StartDate.Before(time.Now()) && f.Filter.EndDate.After(time.Now()) {
			recMap.FieldFilters = append(recMap.FieldFilters, buildFieldFilterItem(f))
		}
	}
	return
}

func buildFieldFilterItem(fieldMappingItem types.FieldMapping) types.FieldFilter {
	valuesMap := make(map[string]bool)
	for _, v := range fieldMappingItem.Filter.Values {
		valuesMap[v] = true
	}
	return types.FieldFilter{
		Index:    fieldMappingItem.Index,
		Path:     fieldMappingItem.RemoteName,
		Criteria: fieldMappingItem.Filter.Criteria,
		Values:   valuesMap,
	}
}

func selectNeastedArrayAddress(previousAddress, canditadeAddress string) (selectedAddress string, err error) {
	lastIndex := strings.LastIndex(canditadeAddress, ".$")
	if lastIndex < 1 {
		selectedAddress = previousAddress
		return
	}
	addressToCheck := canditadeAddress[0:lastIndex]
	if previousAddress == "" {
		selectedAddress = addressToCheck
		return
	}
	if strings.HasPrefix(previousAddress, addressToCheck) {
		selectedAddress = previousAddress
		return
	}
	if strings.HasPrefix(addressToCheck, previousAddress) {
		selectedAddress = addressToCheck
		return
	}
	err = fmt.Errorf(
		"multiple address branches not allowed (%s | %s)",
		previousAddress,
		addressToCheck,
	)
	return
}

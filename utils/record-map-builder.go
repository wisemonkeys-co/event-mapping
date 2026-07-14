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
	for i, fieldMappingItem := range eventConfig.FieldMapping {
		if strings.Contains(fieldMappingItem.RemoteName, ".$") {
			recMap.SplitEventField, err = selectNeastedArrayAddress(recMap.SplitEventField, fieldMappingItem.RemoteName)
			if err != nil {
				err = fmt.Errorf("%s real name %s", err.Error(), realmName)
				return
			}
			eventConfig.FieldMapping[i].RemoteName = strings.Replace(fieldMappingItem.RemoteName, ".$", "", -1)
			fieldMappingItem.RemoteName = eventConfig.FieldMapping[i].RemoteName
		}
		unique := len(fieldMappingItem.Types) == 1
		for _, fieldMappingTypeItem := range fieldMappingItem.Types {
			recFieldItem := types.RecordField{
				Index:      fieldMappingItem.Index,
				Name:       fieldMappingItem.Name,
				RemoteName: fieldMappingItem.RemoteName,
				Format:     fieldMappingItem.Format,
				DataType:   fieldMappingItem.DataType,
				Unique:     unique,
				Expression: fieldMappingItem.Expression,
			}
			switch fieldMappingTypeItem {
			case "id":
				recMap.Id = append(recMap.Id, recFieldItem)
				recMap.IdIndex = append(recMap.IdIndex, recFieldItem)
			case "date":
				recMap.Date = recFieldItem
			case "count":
				recMap.Count = recFieldItem
				recMap.CountIndex = fieldMappingItem.Index
			case "customer":
				recMap.Customer = append(recMap.Customer, recFieldItem)
			case "TPCustomer":
				recMap.TPCustomer = append(recMap.TPCustomer, recFieldItem)
			case "modifier":
				recMap.Modifiers = append(recMap.Modifiers, recFieldItem)
			case "service":
				recMap.Service = append(recMap.Service, recFieldItem)
			case "metric":
				recMap.Metrics = append(recMap.Metrics, recFieldItem)
				recMap.Variables = append(recMap.Variables, recFieldItem)
			case "value":
				recMap.Values = append(recMap.Values, recFieldItem)
				recMap.Variables = append(recMap.Variables, recFieldItem)
			case "info":
				recMap.Info = append(recMap.Info, recFieldItem)
			case "groupby":
				recMap.GroupBy = append(recMap.GroupBy, recFieldItem)
			case "billingScope":
				recMap.BillingScope = append(recMap.BillingScope, recFieldItem)
			}
		}
		if fieldMappingItem.Filter != nil && fieldMappingItem.Filter.StartDate.Before(time.Now()) && fieldMappingItem.Filter.EndDate.After(time.Now()) {
			recMap.FieldFilters = append(recMap.FieldFilters, buildFieldFilterItem(fieldMappingItem))
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

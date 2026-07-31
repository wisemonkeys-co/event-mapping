package types

import "time"

type RecordMap struct {
	Realm           string
	Event           string
	Location        *time.Location
	SplitEventField string
	Id              []RecordField
	IdIndex         []RecordField
	Date            RecordField
	Count           RecordField
	CountIndex      int
	TpName          string
	Customer        []RecordField
	//TPCustomer      []RecordField
	Modifiers []RecordField
	Info      []RecordField
	Service   []RecordField
	Metrics   []RecordField
	//Values          []RecordField
	Variables    []RecordField
	GroupBy      []RecordField
	BillingScope []RecordField
	FieldFilters []FieldFilter
}

type RecordField struct {
	Index      int
	Name       string
	RemoteName string
	Format     string
	Value      string
	DataType   string
	Unique     bool
	Expression string
}

type FieldFilter struct {
	Index    int
	Path     string
	Criteria string
	Values   map[string]bool
}

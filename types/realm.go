package types

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Realm struct {
	ID          bson.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string        `json:"name" bson:"name"`
	Description string        `json:"description" bson:"description"`
	Location    string        `json:"location" bson:"location"`
	SortGroup   string        `json:"sortGroup" bson:"sortGroup"`
	Type        string        `json:"type" bson:"type"`
	Color       string        `json:"color" bson:"color"`
	Events      []Event       `json:"events"`
}

type Event struct {
	Name         string         `json:"name" bson:"name"`
	Description  string         `json:"description" bson:"description"`
	SortIndex    int            `json:"sortIndex" bson:"sortIndex"`
	TpName       string         `json:"tpName,omitempty" bson:"tpName,omitempty"`
	FieldMapping []FieldMapping `json:"fieldMapping" bson:"fieldMapping"`
}

type FieldMapping struct {
	Index       int      `json:"index"`
	Name        string   `json:"name"`
	Description string   `json:"description" bson:"description"`
	RemoteName  string   `json:"remoteName" bson:"remoteName"`
	Types       []string `json:"types"`
	Format      string   `json:"format"`
	Filter      *Filter  `json:"filter" bson:"filter"`
	DataType    string   `json:"dataType" bson:"dataType"`
	Expression  string   `json:"expression" bson:"expression"`
}

type Filter struct {
	Criteria  string    `json:"criteria" bson:"criteria"`
	StartDate time.Time `json:"startDate" bson:"startDate"`
	EndDate   time.Time `json:"endDate" bson:"endDate"`
	Values    []string  `json:"values" bson:"values"`
}

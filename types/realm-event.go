package types

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type RealmEvent struct {
	Realm      string    `json:"realm" bson:"realm"`
	RealmEvent string    `json:"realmEvent" bson:"realmEvent"`
	RealmID    string    `json:"realmId" bson:"realmId"`
	RealmDate  time.Time `json:"realmDate" bson:"realmDate"`
	Count      int32     `json:"count" bson:"count"`
	Customer   []string  `json:"customer" bson:"customer"`
	//TpCustomer         []string               `json:"tpCustomer" bson:"tpCustomer"`
	Metric []map[string]float64 `json:"metric" bson:"metric"`
	//Value              []map[string]float64   `json:"value" bson:"value"`
	Variables    map[string]interface{} `json:"variables"`
	Modifier     []map[string]string    `json:"modifier" bson:"modifier"`
	Service      []string               `json:"service" bson:"service"`
	GroupBy      []string               `json:"groupBy" bson:"groupBy"`
	BillingScope []string               `json:"billingScope" bson:"billingScope"`
	Info         []map[string]string    `json:"info" bson:"info"`
	AccountID    bson.ObjectID          `json:"account,omitempty" bson:"account"`
	//ThirdPartAccountID bson.ObjectID          `json:"tpAccount,omitempty" bson:"tpAccount"`
	TpName      string   `json:"tpName,omitempty" bson:"tpName"`
	MetaData    MetaData `json:"metadata" bson:"metadata"`
	EventSource string   `json:"eventSource" bson:"eventSource"`
	Raw         string   `json:"raw" bson:"raw"`
}

type MetaData struct {
	AccountingId  string    `json:"accountingId" bson:"accountingId"`
	RatingId      string    `json:"ratingId" bson:"ratingId"`
	EventId       string    `json:"eventId" bson:"eventId"`
	Channel       string    `json:"channel" bson:"channel"`
	PipelineStage string    `json:"pipelineStage" bson:"pipelineStage"`
	Status        string    `json:"status" bson:"status"`
	StatusCode    string    `json:"statusCode" bson:"statusCode"`
	Version       int32     `json:"version" bson:"version"`
	Timestamp     time.Time `json:"timestamp" bson:"timestamp"`
}

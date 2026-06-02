package infra

import (
	"context"
	"os"
	"time"

	"github.com/wisemonkeys-co/event-mapping/types"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func FindRealmEventMappingData(client *mongo.Client, realmName string, event string) (eventMappingData types.Event, location string, err error) {
	queryTimeout, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	opts := options.FindOne().
		SetProjection(bson.M{
			"location":              1,
			"events.fieldMapping.$": 1,
			"events.tpName":         1,
		})
	findResult := client.
		Database(os.Getenv("WISE_OCS_DB_NAME")).
		Collection("realms").
		FindOne(queryTimeout, bson.M{
			"name":        realmName,
			"events.name": event,
		}, opts)
	err = findResult.Err()
	if err != nil {
		return
	}
	var realm types.Realm
	err = findResult.Decode(&realm)
	if err != nil {
		return
	}
	eventMappingData = realm.Events[0]
	location = realm.Location
	return
}

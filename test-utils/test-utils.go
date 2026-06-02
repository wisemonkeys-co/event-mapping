package testutils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/wisemonkeys-co/event-mapping/types"
)

// var MongoClient *mongo.Client

func AssertEquals(t *testing.T, actual, expected interface{}) {
	if actual != expected {
		t.Errorf("Expected %v to equal %v", actual, expected)
		t.Log(string(debug.Stack()))
		t.FailNow()
	}
}

// func StartTestDB() error {
// 	if MongoClient != nil {
// 		return nil
// 	}
// 	wiseOcsDbHost := os.Getenv("WISE_OCS_DB_HOST")
// 	if wiseOcsDbHost == "" {
// 		os.Setenv("WISE_OCS_DB_HOST", "mongodb://localhost:27017")
// 	}
// 	os.Setenv("WISE_OCS_DB_NAME", "wiseOCSTest")
// 	var connectError error
// 	initTimeout := 10 * time.Second
// 	opCtx, cancel := context.WithTimeout(context.Background(), initTimeout)
// 	defer cancel()
// 	opts := options.
// 		Client().
// 		ApplyURI(os.Getenv("WISE_OCS_DB_HOST")).
// 		SetWriteConcern(writeconcern.Majority()).
// 		SetConnectTimeout(initTimeout)
// 	MongoClient, connectError = mongo.Connect(opts)
// 	if connectError != nil {
// 		return connectError
// 	}
// 	pingErr := MongoClient.Ping(opCtx, readpref.Primary())
// 	if pingErr != nil {
// 		return pingErr
// 	}
// 	return nil
// }

// func StopTestDB() error {
// 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 	defer cancel()
// 	err := MongoClient.Disconnect(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	MongoClient = nil
// 	return nil
// }

// func InsertOne(collection string, data interface{}) (bson.ObjectID, error) {
// 	testCtx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
// 	defer cancel()
// 	var insertedID bson.ObjectID
// 	inserteOneResult, err := MongoClient.
// 		Database(os.Getenv("WISE_OCS_DB_NAME")).
// 		Collection(collection).
// 		InsertOne(testCtx, data)
// 	if err == nil {
// 		insertedID = inserteOneResult.InsertedID.(bson.ObjectID)
// 	}
// 	return insertedID, err
// }

// func InsertMany(collection string, data []interface{}) ([]bson.ObjectID, error) {
// 	testCtx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
// 	defer cancel()
// 	var insertedID []bson.ObjectID
// 	inserteManyResult, err := MongoClient.
// 		Database(os.Getenv("WISE_OCS_DB_NAME")).
// 		Collection(collection).
// 		InsertMany(testCtx, data)
// 	if err == nil {
// 		for _, id := range inserteManyResult.InsertedIDs {
// 			insertedID = append(insertedID, id.(bson.ObjectID))
// 		}
// 	}
// 	return insertedID, err
// }

// func DeleteDocuments(collection string, filter map[string]interface{}) error {
// 	testCtx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
// 	defer cancel()
// 	_, testError := MongoClient.
// 		Database(os.Getenv("WISE_OCS_DB_NAME")).
// 		Collection(collection).
// 		DeleteMany(testCtx, filter)
// 	return testError
// }

// func DropDatabase() error {
// 	testCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
// 	errorData := MongoClient.
// 		Database(os.Getenv("WISE_OCS_DB_NAME")).
// 		Drop(testCtx)
// 	return errorData
// }

func ReadJsonFiles() (filesData map[string][]interface{}, err error) {
	filesData = make(map[string][]interface{})
	rootPath := GetRootPath()
	files, err := os.ReadDir(rootPath + "/test-utils/json-test-files")
	if err != nil {
		return
	}
	for _, file := range files {
		var fileData []byte
		fileData, err = os.ReadFile(fmt.Sprintf("%s/test-utils/json-test-files/%s", rootPath, file.Name()))
		if err != nil {
			return
		}
		collection := strings.Split(file.Name(), ".")[0]
		filesData[collection], err = decodeJsonToTypes(collection, fileData)
		if err != nil {
			return
		}
	}
	return
}

func GetRootPath() string {
	currentPath, _ := os.Getwd()
	i := strings.LastIndex(currentPath, "/event-mapping")
	return currentPath[:i] + "/event-mapping"
}

// func GetBackupFilesPath() string {
// 	return GetRootPath() + "/test-utils/backup-files/"
// }

// func PopulateTestDB() (err error) {
// 	StartTestDB()
// 	itemsMap, err := ReadJsonFiles()
// 	if err != nil {
// 		return
// 	}
// 	collectionsMap := map[string]bool{
// 		"realms": true,
// 	}
// 	for collection, items := range itemsMap {
// 		if collectionsMap[collection] {
// 			_, err = InsertMany(collection, items)
// 			if err != nil {
// 				return
// 			}
// 		}
// 	}
// 	return
// }

// func CleanTestDB() error {
// 	collections := []string{"realms"}
// 	for _, collection := range collections {
// 		err := DeleteDocuments(collection, map[string]interface{}{})
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

func GetEventDataFromJsonFiles(eventName, realmName string) (eventData types.Event, eventLocation string, err error) {
	files, err := ReadJsonFiles()
	if err != nil {
		return
	}
	for i := 0; i < len(files["realms"]) && eventData.Name == ""; i++ {
		ri := files["realms"][i].(types.Realm)
		for _, e := range ri.Events {
			if e.Name == eventName {
				eventData = e
				eventLocation = ri.Location
				return
			}
		}
	}
	err = errors.New("could not find the requested event data")
	return
}

func decodeJsonToTypes(typeName string, jsonData []byte) (items []interface{}, err error) {
	reader := json.NewDecoder(bytes.NewReader(jsonData))
	switch typeName {
	case "realms":
		var list []types.Realm
		reader.Decode(&list)
		for _, item := range list {
			items = append(items, item)
		}
	case "external-events":
		var list []map[string]interface{}
		reader.Decode(&list)
		for _, item := range list {
			items = append(items, item)
		}
	default:
		err = fmt.Errorf("unknow type %s", typeName)
	}
	return
}

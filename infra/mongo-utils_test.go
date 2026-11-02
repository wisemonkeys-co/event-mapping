package infra

import (
	"context"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Teste de integração: roda contra um MongoDB ou DocumentDB real quando EVENT_MAPPING_TEST_MONGO_URI está definida
// (ex.: mongodb://localhost:27017). Usa um banco temporário, apagado no fim.
func TestFindRealmEventMappingData(t *testing.T) {
	uri := os.Getenv("EVENT_MAPPING_TEST_MONGO_URI")
	if uri == "" {
		t.Skip("EVENT_MAPPING_TEST_MONGO_URI não definida")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(ctx)

	dbName := "event_mapping_test_" + time.Now().Format("20060102150405")
	t.Setenv("WISE_OCS_DB_NAME", dbName)
	db := client.Database(dbName)
	defer db.Drop(ctx)

	// Evento procurado fora da primeira posição: a projeção precisa trazer o evento casado, não o primeiro.
	_, err = db.Collection("realms").InsertOne(ctx, bson.M{
		"name":     "bonuz-br",
		"location": "America/Sao_Paulo",
		"events": bson.A{
			bson.M{"name": "outro", "tpName": "tp-outro", "fieldMapping": bson.A{bson.M{"index": 9, "name": "x"}}},
			bson.M{"name": "daily-burns", "tpName": "tp-burns", "fieldMapping": bson.A{
				bson.M{"index": 0, "name": "id", "types": bson.A{"id"}},
				bson.M{"index": 1, "name": "date", "types": bson.A{"date"}, "format": "2006-01-02"},
			}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	ev, loc, err := FindRealmEventMappingData(client, "bonuz-br", "daily-burns")
	if err != nil {
		t.Fatal(err)
	}
	if loc != "America/Sao_Paulo" || ev.TpName != "tp-burns" || len(ev.FieldMapping) != 2 ||
		ev.FieldMapping[1].Name != "date" || ev.FieldMapping[1].Format != "2006-01-02" {
		t.Errorf("evento = %+v, location = %q", ev, loc)
	}

	if _, _, err := FindRealmEventMappingData(client, "bonuz-br", "nao-existe"); err != mongo.ErrNoDocuments {
		t.Errorf("evento inexistente: err = %v, quer ErrNoDocuments", err)
	}
}

package turbo

import (
	"fmt"
	"labix.org/v2/mgo/bson"
	"testing"
)

var (
	db    *Database
	dbErr error
)

func assertDb() {
	if db != nil {
		return
	}
	db, dbErr = NewDatabase("mongodb://test:test@kahana.mongohq.com:10039/turbodev", "turbodev", "data")
	if dbErr != nil {
		fmt.Errorf("Houston, we have a problem '%s'", dbErr.Error())
	}
}

func TestSet(t *testing.T) {
	assertDb()
	err := db.set("/bransonapp", bson.M{
		"testPath": "hi",
		"testObj": bson.M{
			"one": 2,
			"two": 1,
		},
		"newTestVal": 600,
	})
	if err != nil {
		t.Error(err)
	}
}

func TestGetAll(t *testing.T) {
	assertDb()
	err, result, _ := db.get("/bransonapp")
	if err != nil {
		t.Error(err)
	} else {
		t.Log(result)
	}
}

func TestGet(t *testing.T) {
	assertDb()
	err, result, _ := db.get("/bransonapp/testPath")
	if err != nil {
		t.Error(err)
	} else if result.(string) != "hi" {
		t.Error("should have returned 'hi', actually returned " + result.(string))
	} else {
		t.Log(result)
	}
}

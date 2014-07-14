package turbo

import (
	"labix.org/v2/mgo/bson"
	"testing"
)

func TestGenerateRevUpdate(t *testing.T) {
	revSet := bson.M{}
	database.init("mongodb://bitbeam.info:27017", "test", "entries")
	thing := bson.M{
		"a": bson.M{
			"b": bson.M{
				"c": "hi",
			},
			"d": bson.M{
				"e": "there",
				"f": bson.M{
					"g": 45,
				},
			},
		},
	}
	result := generateRevisionUpdate(thing, "/1/2/3", &revSet)
	if _, _, exists := revSet["/1/2/3/a/b/c"]; !exists {
		t.Error("/1/2/3/a/b/c did not exist", result)
	}
	if _, _, exists := revSet["/1/2/3/a/b"]; !exists {
		t.Error("/1/2/3/a/b did not exist", result)
	}
	if _, _, exists := revSet["/1/2/3/a"]; !exists {
		t.Error("/1/2/3/a did not exist", result)
	}
	if _, _, exists := revSet["/1/2/3/a/d/e"]; !exists {
		t.Error("/1/2/3/a/d/e did not exist", result)
	}
	if _, _, exists := revSet["/1/2/3/a/d"]; !exists {
		t.Error("/1/2/3/a/d did not exist", result)
	}
	if _, _, exists := revSet["/1/2/3/a/d/f/g"]; !exists {
		t.Error("/1/2/3/a/d/f/g did not exist", result)
	}
	if _, _, exists := revSet["/1/2/3/a/d/f"]; !exists {
		t.Error("/1/2/3/a/d/f did not exist", result)
	}
	if _, _, exists := revSet["/1/2/3"]; !exists {
		t.Error("/1/2/3 did not exist", result)
	}
	if _, _, exists := revSet["/1/2"]; !exists {
		t.Error("/1/2 did not exist", result)
	}
	if _, _, exists := revSet["/1"]; !exists {
		t.Error("/1 did not exist", result)
	}
}

func TestSet(t *testing.T) {
	database.init("mongodb://bitbeam.info:27017", "test", "entries")
	err := database.set("/bransonapp", bson.M{"testPath": "hi"})
	if err != nil {
		t.Error(err)
	}
}

func TestGetAll(t *testing.T) {
	database.init("mongodb://bitbeam.info:27017", "test", "entries")
	err, result, _ := database.get("/bransonapp")
	if err != nil {
		t.Error(err)
	} else {
		t.Log(result)
	}
}

func TestGet(t *testing.T) {
	database.init("mongodb://bitbeam.info:27017", "test", "entries")
	err, result, _ := database.get("/bransonapp/testPath")
	if err != nil {
		t.Error(err)
	} else if result.(string) != "hi" {
		t.Error("should have returned 'hi', actually returned " + result.(string))
	} else {
		t.Log(result)
	}
}

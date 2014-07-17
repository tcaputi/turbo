package turbo

import (
	"labix.org/v2/mgo/bson"
	"testing"
)

var (
	db *Database
)

func init() {
	if db != nil return
	db = NewDatabase("mongodb://bitbeam.info:27017", "test", "entries")
}

func TestCompileSetArtifacts(t *testing.T) {
	init()
	revSet := bson.M{}
	database.init("mongodb://bitbeam.info:27017", "test", "entries")
	thing := map[string]interface{}{
		"/1/2/3/": bson.M{
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
	generateRevisionUpdate(thing, "/1/2/3/", &revSet)
	if _, exists := revSet["_rev./1/2/3/a/b/c"]; !exists {
		t.Error("/1/2/3/a/b/c did not exist", revSet)
	}
	if _, exists := revSet["_rev./1/2/3/a/b"]; !exists {
		t.Error("/1/2/3/a/b did not exist", revSet)
	}
	if _, exists := revSet["_rev./1/2/3/a"]; !exists {
		t.Error("/1/2/3/a did not exist", revSet)
	}
	if _, exists := revSet["_rev./1/2/3/a/d/e"]; !exists {
		t.Error("/1/2/3/a/d/e did not exist", revSet)
	}
	if _, exists := revSet["_rev./1/2/3/a/d"]; !exists {
		t.Error("/1/2/3/a/d did not exist", revSet)
	}
	if _, exists := revSet["_rev./1/2/3/a/d/f/g"]; !exists {
		t.Error("/1/2/3/a/d/f/g did not exist", revSet)
	}
	if _, exists := revSet["_rev./1/2/3/a/d/f"]; !exists {
		t.Error("/1/2/3/a/d/f did not exist", revSet)
	}
	if _, exists := revSet["_rev./1/2/3"]; !exists {
		t.Error("/1/2/3 did not exist", revSet)
	}
	if _, exists := revSet["_rev./1/2"]; !exists {
		t.Error("/1/2 did not exist", revSet)
	}
	if _, exists := revSet["_rev./1"]; !exists {
		t.Error("/1 did not exist", revSet)
	}
}

func TestSet(t *testing.T) {
	init()
	database.init("mongodb://bitbeam.info:27017", "test", "entries")
	err := database.set("/bransonapp", bson.M{"testPath": "hi"})
	if err != nil {
		t.Error(err)
	}
}

func TestGetAll(t *testing.T) {
	init()
	database.init("mongodb://bitbeam.info:27017", "test", "entries")
	err, result, _ := database.get("/bransonapp")
	if err != nil {
		t.Error(err)
	} else {
		t.Log(result)
	}
}

func TestGet(t *testing.T) {
	init()
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

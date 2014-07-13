package turbo

import(
	"testing"
	"labix.org/v2/mgo/bson"
)

func TestGenerateRevUpdate(t *testing.T){
	revSet := bson.M{}
	database.init("mongodb://bitbeam.info:27017", "test", "entries")
	thing := bson.M{"a": bson.M{"b": bson.M{"c": "hi"}, "d": bson.M{"e": "there"}}}
	generateRevisionUpdate(thing, "", &revSet)
}

func TestSet(t *testing.T){
	database.init("mongodb://bitbeam.info:27017", "test", "entries")
	err := database.set("/bransonapp", bson.M{"testPath": "hi"})
	if err != nil {
		t.Error(err)
	}
}

func TestGetAll(t *testing.T){
	database.init("mongodb://bitbeam.info:27017", "test", "entries")
	err, result, _ := database.get("/bransonapp")
	if err != nil {
		t.Error(err)
	}else{
		t.Log(result);
	}
}

func TestGet(t *testing.T){
	database.init("mongodb://bitbeam.info:27017", "test", "entries")
	err, result, _ := database.get("/bransonapp/testPath")
	if err != nil {
		t.Error(err)
	}else if result.(string) != "hi" {
		t.Error("should have returned 'hi', actually returned " + result.(string))
	}else{
		t.Log(result)
	}
}
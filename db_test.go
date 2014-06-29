package turbo

import(
	"testing"
	"labix.org/v2/mgo/bson"
)

func TestSet(t *testing.T){
	database.init("mongodb://bitbeam.info:27017", "test", "entries")
	err := database.set("bransonapp", bson.M{"testPath": "hi"})
	if err != nil {
		t.Error(err)
	}
}

func TestGet(t *testing.T){
	database.init("mongodb://bitbeam.info:27017", "test", "entries")
	err, result := database.get("bransonapp")
	if err != nil {
		t.Error(err)
	}else{
		t.Log(result);
	}
}
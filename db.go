package turbo

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strings"
)

type Database struct {
	client *mgo.Session
	col	*mgo.Collection
}

var(
	database = &Database{}
)

func unwrapValue(path string, object interface{}) interface{}{
	pathStrings := strings.Split(path, ".")
	for i:=0; i<len(pathStrings); i++ {
		object = object.(bson.M)[pathStrings[i]]
	}
	return object
}

func generateRevisionUpdate(obj interface{}, basePath string, revSet *bson.M){
	if basePath == "/"{
		basePath = ""
	}
	if _, ok := obj.(bson.M); ok{
		for key, value := range obj.(bson.M) {
			generateRevisionUpdate(value, basePath + "/" + key, revSet)
		}
	}
	(*revSet)["_rev." + basePath] = 0
}


func (db *Database) init(mgoPath string, dbName string, collectionName string){
	session, err := mgo.Dial(mgoPath)
	if err != nil {
		panic(err)
		return
	}
	db.client = session
	db.col = session.DB(dbName).C(collectionName)
}

func (db *Database) get(path string) (error, interface{}, int){
	var result bson.M
	dotPath := "_tree" + strings.Replace(path, "/", ".", -1)
	revPath := "_rev." + path
	err := db.col.Find(nil).Select(bson.M{dotPath: 1, revPath: 1}).One(&result)
	if err != nil {
		return err, nil, 0
	}else{
		return nil, unwrapValue(dotPath, result), result["_rev"].(bson.M)[path].(int)
	}
}

func (db *Database) set(path string, value interface{}) error{
	dotPath := "_tree" + strings.Replace(path, "/", ".", -1)
	revUpdate := bson.M{}
	generateRevisionUpdate(value, path, &revUpdate)
	update:= bson.M{"$set": bson.M{dotPath: value}, "$inc": revUpdate}
	_, err := db.col.Upsert(nil, update)
	return err
}
package turbo

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strings"
)

type Database struct {
	client *mgo.Session
	col    *mgo.Collection
}

var (
	database = &Database{}
)

func unwrapValue(path string, object interface{}) interface{} {
	pathStrings := strings.Split(path, ".")
	for i := 0; i < len(pathStrings); i++ {
		object = object.(bson.M)[pathStrings[i]]
	}
	return object
}

func generateRevisionUpdateHelper(obj interface{}, basePath string, revSet *bson.M) {
	// Sanitize path
	if _, ok := obj.(bson.M); ok {
		var subPath string
		for key, value := range obj.(bson.M) {
			if strings.HasSuffix(basePath, "/") {
				subPath = basePath + key
			} else {
				subPath = basePath + "/" + key
			}

			generateRevisionUpdateHelper(value, subPath, revSet)
		}
	}

	(*revSet)["_rev."+basePath] = 0
}

func generateRevisionUpdate(obj interface{}, path string) {
	revSet := bson.M{}
	generateRevisionUpdateHelper(obj, path, &revSet)
	// IWASHERE
	path =
}

func (db *Database) init(mgoPath string, dbName string, collectionName string) {
	session, err := mgo.Dial(mgoPath)
	if err != nil {
		panic(err)
		return
	}
	db.client = session
	db.col = session.DB(dbName).C(collectionName)
}

func (db *Database) get(path string) (error, interface{}, int) {
	var result bson.M
	dotPath := "_tree" + strings.Replace(path, "/", ".", -1)
	revPath := "_rev." + path
	err := db.col.Find(nil).Select(bson.M{dotPath: 1, revPath: 1}).One(&result)
	if err != nil {
		return err, nil, 0
	} else {
		return nil, unwrapValue(dotPath, result), result["_rev"].(bson.M)[path].(int)
	}
}

func (db *Database) set(path string, value interface{}) error {
	dotPath := "_tree" + strings.Replace(path, "/", ".", -1)
	revUpdate := bson.M{}
	generateRevisionUpdate(value, path, &revUpdate)
	update := bson.M{"$set": bson.M{dotPath: value}, "$inc": revUpdate}
	_, err := db.col.Upsert(nil, update)
	return err
}

package turbo

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strings"
)

const (
	DB_REV_NODE      = "_rev"
	DB_TREE_NODE     = "_tree"
	DB_REV_INCREMENT = 1
	DB_SET           = "$set"
	DB_INC           = "$inc"
)

type Database struct {
	client *mgo.Session
	col    *mgo.Collection
}

func NewDatabase(mgoPath string, dbName string, collectionName string) (*Database, error) {
	session, err := mgo.Dial(mgoPath)
	if err != nil {
		return nil, err
	}
	db := Database{}
	db.client = session
	db.col = session.DB(dbName).C(collectionName)
	return &db, nil
}

func (db *Database) unwrapValue(path string, object interface{}) interface{} {
	pathStrings := strings.Split(path, DOT)
	for i := 0; i < len(pathStrings); i++ {
		object = object.(bson.M)[pathStrings[i]]
	}
	return object
}

func (db *Database) compileSetArtifacts(obj interface{}, path string) (bson.M, bson.M) {
	revMap := bson.M{}
	setMap := bson.M{}

	if _, ok := obj.(map[string]interface{}); ok {
		var subPath string
		for key, value := range obj.(map[string]interface{}) {
			revMap[(DB_REV_NODE + DOT + key)] = DB_REV_INCREMENT
			setMap[(DB_TREE_NODE + DOT + mongoizePath(key))] = value
		}
		// Handle parent segments of path for rev set
		for strings.LastIndex(path, SLASH) > 0 {
			path = path[:strings.LastIndex(path, SLASH)]
			revMap[(DB_REV_NODE + DOT + path)] = DB_REV_INCREMENT
		}
	}

	return revMap, setMap
}

func (db *Database) get(path string) (error, interface{}, int) {
	var result bson.M
	dotPath := DB_TREE_NODE + DOT + mongoizePath(path)
	revPath := DB_REV_NODE + path

	err := db.col.Find(nil).Select(bson.M{
		dotPath: 1,
		revPath: 1,
	}).One(&result)
	if err != nil {
		return err, nil, -1
	} else {
		return nil, db.unwrapValue(dotPath, result), result[DB_REV_NODE].(bson.M)[path].(int)
	}
}

func (db *Database) set(path string, value interface{}) error {
	revMap, setMap := db.compileSetArtifacts(value, path)
	_, err := db.col.Upsert(nil, bson.M{
		DB_SET: setMap,
		DB_INC: revMap,
	})
	return err
}

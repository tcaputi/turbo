package turbo

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strings"
)

const (
	DB_REV_NODE      = "_rev"
	DB_REL_NODE      = "_rel"
	DB_TREE_NODE     = "_tree"
	DB_REV_INCREMENT = 1
	DB_SET           = "$set"
	DB_SET_ONCE      = "$setOnInsert"
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

func (db *Database) traceDataRecursive(data interface{}, path string, rev *bson.M) {
	// First check wtf data is
	dataMap, isMap := data.(map[string]interface{})
	dataBsonM, isBsonM := data.(bson.M)
	// If its a pseudo obj type, iterate over its fields
	if isMap {
		for key, val := range dataMap {
			db.traceDataRecursive(val, joinPaths(path, key), rev)
		}
	} else if isBsonM {
		for key, val := range dataBsonM {
			db.traceDataRecursive(val, joinPaths(path, key), rev)
		}
	}
	// Add this to the rev
	(*rev)[(DB_REV_NODE + DOT + path)] = 1
}

func (db *Database) traceData(data interface{}, path string) *bson.M {
	rev := bson.M{}
	// Traverse obj to build rev
	db.traceDataRecursive(data, path, &rev)
	// Now traverse upwards
	for {
		parentPath, hasParentPath := parentOf(path)
		if !hasParentPath {
			break
		} else {
			path = parentPath
			rev[(DB_REV_NODE + DOT + path)] = 1
		}
	}
	return &rev
}

func (db *Database) get(path string) (error, interface{}, int) {
	var result bson.M
	dotPath := DB_TREE_NODE + DOT + mongoizePath(path)
	revPath := DB_REV_NODE + DOT + path
	query := bson.M{}
	// Set the query paths
	query[dotPath] = 1
	query[revPath] = 1

	err := db.col.Find(nil).Select(query).One(&result)
	if err != nil {
		return err, nil, -1
	} else {
		return nil, db.unwrapValue(dotPath, result), result[DB_REV_NODE].(bson.M)[path].(int)
	}
}

func (db *Database) set(path string, data interface{}) error {
	rev := db.traceData(data, path)
	query := bson.M{}
	// Set the query paths
	query[DB_SET] = bson.M{
		(DB_TREE_NODE + DOT + mongoizePath(path)): data,
	}
	query[DB_INC] = *rev
	_, err := db.col.Upsert(nil, query)
	return err
}

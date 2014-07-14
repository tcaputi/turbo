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

// DB Structure
// ============
//
// - _tree
// 		- a
//			- 1: 'value'
//			- 5: 'value'
// 		- b
//			- 2: 'value'
//			- 6: 'value'
// 		- c
//			- 3: 'value'
//			- 7: 'value'
// 		- d
//			- 4: 'value'
//			- 8: 'value'
//
// - _rev
// 		- /a/1:	rev1
// 		- /a/5:	rev5
// 		- /b/2:	rev2
// 		- /b/6:	rev6
// 		- /c/3:	rev3
// 		- /c/7:	rev7
// 		- /d/4:	rev4
// 		- /d/8:	rev8
//
// Query Algos
// ===========
//
// Get
// ---
// Params: path
// Procedure:
//		- Look for _meta.{{path}}
//			- If !exists return nil
//			- If exists, get {{treeId}}
//				- Find subdoc where _id is {{treeId}}
//				- Return value of subdoc
//
// Set
// ---
// Params: path, value
// Procedure:
//		- Look for _meta.{{path}}
//			- If !exists return nil
//			- If exists, get {{treeId}}
//				- Find subdoc where _id is {{treeId}}
//				- Return value of subdoc

func unwrapValue(path string, object interface{}) interface{} {
	pathStrings := strings.Split(path, "/")
	for i := 0; i < len(pathStrings); i++ {
		object = object.(bson.M)[pathStrings[i]]
	}
	return object
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

func (db *Database) get(path string) (error, interface{}) {
	var result bson.M
	err := db.col.Find(nil).Select(bson.M{path: 1}).One(&result)
	if err != nil {
		return err, nil
	} else {
		return nil, unwrapValue(path, result)
	}
}

func (db *Database) set(path string, value interface{}) error {
	update := bson.M{"$set": bson.M{path: value}}
	_, err := db.col.Upsert(nil, update)
	return err
}

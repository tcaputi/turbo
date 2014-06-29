package turbo

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type Database struct {
	client *mgo.Session
	col	*mgo.Collection
}

var(
	database = &Database{}
)

func (db *Database) init(mgoPath string, dbName string, collectionName string){
	session, err := mgo.Dial(mgoPath)
	if err != nil {
		panic(err)
		return
	}
	db.client = session
	db.col = session.DB(dbName).C(collectionName)
}

func (db *Database) get(path string) (error, interface{}){
	var result interface{}
	err := db.col.Find(nil).Select(bson.M{path: 1}).One(&result)
	return err, result
}

func (db *Database) set(path string, value interface{}) error{
	update:= bson.M{"$set": bson.M{(path: value}}
	_, err := db.col.Upsert(nil, update)
	return err
}

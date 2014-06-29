package turbo

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type Entry struct {
	ID    bson.ObjectId `bson:"_id,omitempty"`
	Path  string
	Value interface{}
	Revision int
}

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

func (db *Database) set(path string, value interface{}) (error, bool){
	query := bson.M{"path": path}
	change := mgo.Change{
		Update: bson.M{"$inc": bson.M{"revision": 1}, "$set": bson.M{"value": value}},
		Upsert: true,
	}
	doc := Entry{}
	info, err := db.col.Find(query).Apply(change, &doc)
	return err, (info.Updated == 0)
}

func (db *Database) get(path string) (error, interface{}){
	result := Entry{}
	err := db.col.Find(bson.M{"path": path}).Select(bson.M{"value": 1}).One(&result)
	return err, result.Value
}
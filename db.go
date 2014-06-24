package turbo

import (
	"errors"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type Entry struct {
	Id       bson.ObjectId `bson:"_id,omitempty"`
	Path     string
	Revision int
	Value    []byte
}

type DB struct {
	mgoClient *mgo.Session
	c         *mgo.Collection
}

func (db *DB) init(url string, database string, collection string) error {
	client, err := mgo.Dial(url)
	db.mgoClient = client
	db.c = db.mgoClient.DB(database).C(collection)
	return err
}

func (db *DB) get(path string) (*[]byte, error) {
	result := Entry{}
	err := db.c.Find(bson.M{"Path": path}).One(&result)
	return &(result.Value), err
}

func (db *DB) getTransaction(path string) (*[]byte, int, error) {
	result := Entry{}
	err := db.c.Find(bson.M{"Path": path}).One(&result)
	return &(result.Value), result.Revision, err
}

func (db *DB) set(path string, value *[]byte) error {
	result := Entry{}
	change := mgo.Change{
		Update: bson.M{"$inc": bson.M{"Revision": 1}, "value": value},
	}
	_, err := db.c.Find(bson.M{"Path": path}).Apply(change, &result)
	return err
}

func (db *DB) setTransaction(path string, revision int, value *[]byte) error {
	value, rev, err := db.getTransaction(path)
	if err != nil {
		return err
	}
	if rev == revision {
		return db.set(path, value)
	} else {
		return errors.New("conflict")
	}
}

func (db *DB) delete(path string) error {
	return db.c.Remove(bson.M{"Path": path})
}

func (db *DB) close() {
	db.mgoClient.Close()
}

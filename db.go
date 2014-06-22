package main

import (
	//    "labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

type Entry struct {
	Id        bson.ObjectId `bson:"_id,omitempty"`
	Path      string
	Revision  int
	Value     interface{}
	Timestamp time.Time
}

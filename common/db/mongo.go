package db

import (
	"errors"
	"fmt"
	"time"

	"gopkg.in/mgo.v2"
)

type MongoDB struct {
	Session *mgo.Session
	DBName  string
}

func InitMongo(connectUrl string, dbName string, timeout time.Duration) (Mongo *MongoDB, err error) {
	var session *mgo.Session
	Mongo = &MongoDB{}
	session, err = mgo.DialWithTimeout(connectUrl, timeout*time.Second)
	if err != nil {
		err = errors.New(fmt.Sprintf("Could not connect to %s: %s.", connectUrl, err.Error()))
		return
	}
	session.SetMode(mgo.Strong, true)
	session.SetPoolLimit(100)
	Mongo.Session = session
	Mongo.DBName = dbName
	return
}

func (self *MongoDB) Close() {
	if self.Session != nil {
		self.Session.Close()
	}
}

func (self *MongoDB) C(collectionName string) *mgo.Collection {
	db := self.Session.DB(self.DBName)
	return db.C(collectionName)
}

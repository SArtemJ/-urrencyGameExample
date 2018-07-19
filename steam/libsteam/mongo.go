package libsteam

import (
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MongoStorage struct {
	Session    *mgo.Session
	Db         *mgo.Database
	DbName     string
	Collection string
}

type GameInfo struct {
	ID   bson.ObjectId `bson:"_id" json:"id"`
	Name string        `bson:"name" json:"name"`
	Cost float64       `bson:"cost" json:"cost"`
}

func NewMongoStorage(uri string, databaseName string) *MongoStorage {
	session, err := mgo.Dial(uri)
	if err != nil {
		Logger.Fatalw("Failer connect to MongoDB server",
			"uri", uri,
			"error", err,
		)
	}
	session.SetPoolLimit(200)
	s := &MongoStorage{
		Session:    session,
		DbName:     databaseName,
		Db:         session.DB(databaseName),
		Collection: "Games",
	}
	Logger.Debugw("Connected to MongoDB server",
		"uri", uri,
		"database", databaseName,
	)

	return s
}

func (s *MongoStorage) Close() {
	s.Session.Close()
	s.Session = nil
	s.Db = nil
}

func (s MongoStorage) Reset() {
	s.Db.C(s.Collection).RemoveAll(nil)
}

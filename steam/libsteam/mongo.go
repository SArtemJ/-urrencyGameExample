package libsteam

import (
	"strconv"
	"sync"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MongoStorage struct {
	Session    *mgo.Session
	Db         *mgo.Database
	DbName     string
	Collection string
}

type AppsStruct struct {
	ID    bson.ObjectId `bson:"_id" json:"id"`
	Appid int           `bson:"appid" json:"appid"`
	Name  string        `bson:"name" json:"name"`
	USD   int           `bson:"USD" json:"USD"`
	EUR   int           `bson:"EUR" json:"EUR"`
	GBP   int           `bson:"GBP" json:"GBP"`
	RUB   int           `bson:"RUB" json:"RUB"`
	BTC   int           `bson:"BTC" json:"BTC"`
}

type ApplistStruct struct {
	Apps []AppsStruct `json:"apps"`
}

type SteamApps struct {
	Applist ApplistStruct `json:"applist"`
}

type SteamAppPrice struct {
	Success bool
	Data    struct {
		Price AppPrice `json:"price_overview"`
	} `json:"data"`
}

type AppPrice struct {
	Currency string
	Initial  int
	Final    int
	Discount int `json:"discount_percent"`
}

type AppsWithMutex struct {
	App AppsStruct
	M   sync.RWMutex
}

// type PriceOverview struct {
// 	Currency        string `json:"currency"`
// 	Initial         int    `json:"initial"`
// 	Final           int    `json:"final"`
// 	DiscountPercent int    `json:"discount_percent"`
// }

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

func (s MongoStorage) CheckAndReturnGameInDB(appid string) (AppsWithMutex, bool) {
	var app AppsWithMutex
	appID, err := strconv.Atoi(appid)
	if err != nil {
		Logger.Debugw("Bad id to request try again", err)
		return app, false
	}

	if err := s.Db.C(s.Collection).Find(bson.M{"appid": appID}).One(&app.App); err != nil {
		Logger.Debugw("Can't find app in databse with", " id - ", appid)
		return app, false
	}
	return app, true
}

func (s MongoStorage) UpdateFiledByID(appMongoID bson.ObjectId, field string, value interface{}) bool {
	err := s.Db.C(s.Collection).Update(bson.M{"_id": appMongoID}, bson.M{"$set": bson.M{field: value}})
	if err != nil {
		Logger.Debugw("Can't save game cost USD in mongo", err)
		return false
	}
	Logger.Debugw("Game from DB update success", " gameID - ", appMongoID)
	return true
}

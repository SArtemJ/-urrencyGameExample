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
	USD   float64       `bson:"USD" json:"USD"`
	EUR   float64       `bson:"EUR" json:"EUR"`
	GBP   float64       `bson:"GBP" json:"GBP"`
	RUB   float64       `bson:"RUB" json:"RUB"`
	BTC   float64       `bson:"BTC" json:"BTC"`
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
	M   sync.Mutex
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

/*Reset
очищаем БД
*/
func (s MongoStorage) Reset() {
	s.Db.C(s.Collection).RemoveAll(nil)
}

/*
CheckAndReturnGameInDB
проверяет наличие игры с указанным id в базе Mongo
если игра есть возвращает ее для дальнейшей работы
appid - id игры соответсвует appid из базы Steam
*/
func (s MongoStorage) CheckAndReturnGameInDB(appid string) (*AppsWithMutex, bool) {
	var app AppsWithMutex
	appID, err := strconv.Atoi(appid)
	if err != nil {
		Logger.Debugw("Bad id to request try again", err)
		return &app, false
	}

	if err := s.Db.C(s.Collection).Find(bson.M{"appid": appID}).One(&app.App); err != nil {
		Logger.Debugw("Can't find app in databse with", " id - ", appid)
		return &app, false
	}
	return &app, true
}

/*UpdateFiledByID
обновляет значение поля в MongoDB по соответсвующему ID игры
appMongoID - id записи в базе Mongo
field - поле которое обновляем
value - значение которым обновляем
*/
func (s MongoStorage) UpdateFiledByID(appMongoID bson.ObjectId, field string, value interface{}) bool {
	err := s.Db.C(s.Collection).Update(bson.M{"_id": appMongoID}, bson.M{"$set": bson.M{field: value}})
	if err != nil {
		Logger.Debugw("Can't save game cost USD in mongo", err)
		return false
	}
	Logger.Debugw("Game from DB update success", " gameID - ", appMongoID)
	return true
}

package main

import (
	"zmq4"
	"fmt"
	"time"
	"encoding/json"
	"gopkg.in/mgo.v2"
)

const DB_NAME = "crawl"
const C_NAME = "malshare"

type HashData struct {
	Hash    string    `json:"hash" bson:"hash"`
	Type    string    `json:"type" bson:"type"`
	Created time.Time `json:"created" bson:"created"`
	Desc    string    `json:"desc" bson:"desc"`
}

func saveToDB(collection *mgo.Collection, jsonData []byte) {
	var data HashData

	e := json.Unmarshal(jsonData, &data)
	if e != nil {
		return
	}
	insertData(collection, data)

}

func getSession() (session *mgo.Session, err error) {
	session, err = mgo.Dial("localhost")
	if err != nil {
		fmt.Println("Can't connect to db")
		return nil, err
	}

	return session, nil
}
func getCollections(session *mgo.Session) *mgo.Collection {
	return session.DB(DB_NAME).C(C_NAME)
}

func insertData(collection *mgo.Collection, data HashData) {
	fmt.Println(data)
	e := collection.Insert(data)
	if e != nil {
		fmt.Println(e.Error())
	}
}

func main() {
	s, e := getSession()
	if e != nil {
		fmt.Println(e.Error())
		return
	}
	c := getCollections(s)

	context, _ := zmq4.NewContext()

	sub, _ := context.NewSocket(zmq4.SUB)

	sub.SetSubscribe("")

	sub.Bind("tcp://*:5555")

	for {
		obj, _ := sub.RecvBytes(0)
		fmt.Println(obj)
		saveToDB(c, obj)
	}

}

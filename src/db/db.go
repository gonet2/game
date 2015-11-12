package db

import (
	"gopkg.in/mgo.v2"
	"log"
	"os"
	"time"
)

const (
	DEFAULT_MGO_TIMEOUT = 300
	DEFAULT_CONCURRENT  = 128
	DEFAULT_MONGODB_URL = "mongodb://172.17.42.1/mydb"
	ENV_MONGODB         = "MONGODB_URL"
)

type Database struct {
	session *mgo.Session
	latch   chan *mgo.Session
}

func (db *Database) Init() {
	// create latch
	db.latch = make(chan *mgo.Session, DEFAULT_CONCURRENT)
	// connect db
	mongodb_url := DEFAULT_MONGODB_URL
	if env := os.Getenv(ENV_MONGODB); env != "" {
		mongodb_url = env
	}
	sess, err := mgo.Dial(mongodb_url)
	if err != nil {
		log.Println("mongodb: cannot connect to", os.Getenv(mongodb_url), err)
		os.Exit(-1)
	}

	// set params
	sess.SetMode(mgo.Strong, true)
	sess.SetSocketTimeout(DEFAULT_MGO_TIMEOUT * time.Second)
	sess.SetCursorTimeout(0)
	db.session = sess

	for k := 0; k < cap(db.latch); k++ {
		db.latch <- sess.Copy()
	}
}

func (db *Database) Execute(f func(sess *mgo.Session) error) error {
	// latch control
	sess := <-db.latch
	defer func() {
		db.latch <- sess
	}()

	return f(sess)
}

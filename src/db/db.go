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
	sess, err := mgo.Dial(os.Getenv(ENV_MONGODB))
	if err != nil {
		log.Println("mongodb: cannot connect to", os.Getenv(ENV_MONGODB), err)
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

func (db *Database) Execute(f func(sess *mgo.Session)) {
	// latch control
	sess := <-db.latch
	defer func() {
		db.latch <- sess
	}()

	f(sess)
}

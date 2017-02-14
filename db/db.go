package db

import (
	"os"
	"time"

	log "github.com/Sirupsen/logrus"

	mgo "gopkg.in/mgo.v2"
)

type Database struct {
	session *mgo.Session
	latch   chan *mgo.Session
}

func (db *Database) Init(addr string, concurrent int, timeout time.Duration) {
	// create latch
	db.latch = make(chan *mgo.Session, concurrent)
	sess, err := mgo.Dial(addr)
	if err != nil {
		log.Println("mongodb: cannot connect to - ", addr, err)
		os.Exit(-1)
	}

	// set params
	sess.SetMode(mgo.Strong, true)
	sess.SetSocketTimeout(timeout)
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
	sess.Refresh()
	return f(sess)
}

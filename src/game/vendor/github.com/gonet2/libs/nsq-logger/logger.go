package nsqlogger

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/pquerna/ffjson/ffjson"
)

const (
	ENV_NSQD         = "NSQD_HOST"
	DEFAULT_PUB_ADDR = "http://172.17.42.1:4151/mpub?topic=LOG&binary=true"
	MIME             = "application/octet-stream"
)

const (
	FINEST byte = iota
	FINE
	DEBUG
	TRACE
	INFO
	WARNING
	ERROR
	CRITICAL
)

// log format to sent to nsqd, packed with json
type LogFormat struct {
	Prefix string
	Time   time.Time
	Host   string
	Level  byte
	Msg    string
	Caller string
	LineNo int
}

var (
	_pub_addr string
	_prefix   string
	_level    byte
	_ch       chan []byte
	_flush    chan chan struct{}
)

func init() {
	// get nsqd publish address
	_pub_addr = DEFAULT_PUB_ADDR
	if env := os.Getenv(ENV_NSQD); env != "" {
		_pub_addr = env + "/mpub?topic=LOG&binary=true"
	}
	_ch = make(chan []byte, 4096)
	_flush = make(chan chan struct{}, 1)
	go publish_task()
}

// aggregate & mpub
func publish_task() {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	size := make([]byte, 4)

	flush := func() {
		n := len(_ch)
		if n == 0 {
			return
		}
		// [ 4-byte num messages ]
		// [ 4-byte message #1 size ][ N-byte binary data ]
		//    ... (repeated <num_messages> times)
		buf := new(bytes.Buffer)
		binary.BigEndian.PutUint32(size, uint32(n))
		buf.Write(size)
		for i := 0; i < n; i++ {
			bts := <-_ch
			binary.BigEndian.PutUint32(size, uint32(len(bts)))
			buf.Write(size)
			buf.Write(bts)
		}

		// http part
		resp, err := http.Post(_pub_addr, MIME, buf)
		if err != nil {
			log.Println(err)
			return
		}
		if _, err := ioutil.ReadAll(resp.Body); err != nil {
			log.Println(err)
		}
		resp.Body.Close()
	}

	for {
		select {
		case <-ticker.C:
			flush()
		case w := <-_flush:
			flush()
			w <- struct{}{}
		}
	}
}

// publish to nsqd (localhost nsqd is suggested!)
func publish(msg LogFormat) {
	// fill in the common fields
	hostname, _ := os.Hostname()
	msg.Host = hostname
	msg.Time = time.Now()
	msg.Prefix = _prefix

	// Determine caller func
	if pc, _, lineno, ok := runtime.Caller(2); ok {
		msg.Caller = runtime.FuncForPC(pc).Name()
		msg.LineNo = lineno
	}

	// pack message
	if bts, err := ffjson.Marshal(msg); err == nil {
		_ch <- bts
	} else {
		log.Println(err, msg)
		return
	}
}

// flush remaining logs
func Flush() {
	w := make(chan struct{}, 1)
	_flush <- w
	<-w
}

// set prefix
func SetPrefix(prefix string) {
	_prefix = prefix
}

// set loglevel
func SetLogLevel(level byte) {
	_level = level
}

// wrappers for diffent loglevels
func Finest(v ...interface{}) {
	if _level <= FINEST {
		msg := LogFormat{Level: FINEST, Msg: fmt.Sprint(v...)}
		publish(msg)
	}
}

func Finestf(format string, v ...interface{}) {
	if _level <= FINEST {
		msg := LogFormat{Level: FINEST, Msg: fmt.Sprintf(format, v...)}
		publish(msg)
	}
}

func Fine(v ...interface{}) {
	if _level <= FINE {
		msg := LogFormat{Level: FINE, Msg: fmt.Sprint(v...)}
		publish(msg)
	}
}

func Finef(format string, v ...interface{}) {
	if _level <= FINE {
		msg := LogFormat{Level: FINE, Msg: fmt.Sprintf(format, v...)}
		publish(msg)
	}
}

func Debug(v ...interface{}) {
	if _level <= DEBUG {
		msg := LogFormat{Level: DEBUG, Msg: fmt.Sprint(v...)}
		publish(msg)
	}
}

func Debugf(format string, v ...interface{}) {
	if _level <= DEBUG {
		msg := LogFormat{Level: DEBUG, Msg: fmt.Sprintf(format, v...)}
		publish(msg)
	}
}

func Trace(v ...interface{}) {
	if _level <= TRACE {
		msg := LogFormat{Level: TRACE, Msg: fmt.Sprint(v...)}
		publish(msg)
	}
}

func Tracef(format string, v ...interface{}) {
	if _level <= TRACE {
		msg := LogFormat{Level: TRACE, Msg: fmt.Sprintf(format, v...)}
		publish(msg)
	}
}

func Info(v ...interface{}) {
	if _level <= INFO {
		msg := LogFormat{Level: INFO, Msg: fmt.Sprint(v...)}
		publish(msg)
	}
}

func Infof(format string, v ...interface{}) {
	if _level <= INFO {
		msg := LogFormat{Level: INFO, Msg: fmt.Sprintf(format, v...)}
		publish(msg)
	}
}

func Warning(v ...interface{}) {
	if _level <= WARNING {
		msg := LogFormat{Level: WARNING, Msg: fmt.Sprint(v...)}
		publish(msg)
	}
}

func Warningf(format string, v ...interface{}) {
	if _level <= WARNING {
		msg := LogFormat{Level: WARNING, Msg: fmt.Sprintf(format, v...)}
		publish(msg)
	}
}

func Error(v ...interface{}) {
	if _level <= ERROR {
		msg := LogFormat{Level: ERROR, Msg: fmt.Sprint(v...)}
		publish(msg)
	}
}

func Errorf(format string, v ...interface{}) {
	if _level <= ERROR {
		msg := LogFormat{Level: ERROR, Msg: fmt.Sprintf(format, v...)}
		publish(msg)
	}
}

func Critical(v ...interface{}) {
	if _level <= CRITICAL {
		msg := LogFormat{Level: CRITICAL, Msg: fmt.Sprint(v...)}
		publish(msg)
	}
}

func Criticalf(format string, v ...interface{}) {
	if _level <= CRITICAL {
		msg := LogFormat{Level: CRITICAL, Msg: fmt.Sprintf(format, v...)}
		publish(msg)
	}
}

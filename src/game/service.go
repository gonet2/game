package main

import (
	"errors"
	log "github.com/gonet2/libs/nsq-logger"
	"io"
	"sync"
)

import (
	"client_handler"
	"misc/packet"
	. "proto"
	"registry"
	. "types"
)

const (
	SERVICE = "[GAME]"
)

var (
	ERROR_INCORRECT_FRAME_TYPE = errors.New("incorrect frame type")
	ERROR_SERVICE_NOT_BIND     = errors.New("service not bind")
	ERROR_USER_NOT_REGISTERED  = errors.New("user not registered")
)

type server struct {
	sync.Mutex
}

func (s *server) latch(f func()) {
	s.Lock()
	defer s.Unlock()
	f()
}

func (s *server) Stream(stream GameService_StreamServer) error {
	var sess Session
	for {
		in, err := stream.Recv()
		if err == io.EOF { // client closed
			return nil
		}

		if err != nil {
			log.Critical("recv err", err)
			return err
		}

		switch in.Type {
		case Game_Message:
			// flag validation
			if sess.Flag&SESS_REGISTERED == 0 {
				log.Critical("user not registered")
				return ERROR_USER_NOT_REGISTERED
			}

			// locate handler
			reader := packet.Reader(in.Message)
			c, err := reader.ReadS16()
			if err != nil {
				log.Critical(err)
				return err
			}
			handle := client_handler.Handlers[c]
			if handle == nil {
				log.Criticalf("service not bind: %v", c)
				return ERROR_SERVICE_NOT_BIND

			}

			// process
			var ret []byte
			wrap := func() { ret = handle(&sess, reader) }
			s.latch(wrap)

			// write return value
			if ret != nil {
				if err := stream.Send(&Game_Frame{Type: Game_Message, Message: ret}); err != nil {
					log.Critical(err)
					return err
				}
			}

			// session control
			if sess.Flag&SESS_KICKED_OUT != 0 { // logic kick out
				if err := stream.Send(&Game_Frame{Type: Game_Kick}); err != nil {
					log.Critical(err)
					return err
				}
				return nil
			}
		case Game_Register:
			// TODO: create session
			sess.Flag |= SESS_REGISTERED
			registry.Register(sess.UserId, &stream)
			log.Trace("user registered")
		case Game_Unregister:
			// TODO: destroy session & return
			registry.Unregister(sess.UserId)
			log.Trace("user unregistered")
			return nil
		default:
			log.Errorf("incorrect frame type: %v", in.Type)
			return ERROR_INCORRECT_FRAME_TYPE
		}
	}
}

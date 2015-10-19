package main

import (
	"errors"
	"io"
	"sync"
	"time"

	log "github.com/gonet2/libs/nsq-logger"
)

import (
	"client_handler"
	"misc/packet"
	. "proto"
	"registry"
	. "types"
)

const (
	_port = ":51000"
)

const (
	SERVICE      = "[GAME]"
	RECV_TIMEOUT = 5 * time.Second
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

// stream receiver
func (s *server) recv(stream GameService_StreamServer, sess_die chan struct{}) chan *Game_Frame {
	ch := make(chan *Game_Frame, 1)
	go func() {
		defer func() {
			close(ch)
		}()
		for {
			in, err := stream.Recv()
			if err == io.EOF { // client closed
				return
			}

			if err != nil {
				log.Critical(err)
				return
			}
			select {
			case ch <- in:
			case <-sess_die:
			}
		}
	}()
	return ch
}

// stream server
func (s *server) Stream(stream GameService_StreamServer) error {
	// session init
	var sess Session
	sess_die := make(chan struct{})
	ch_agent := s.recv(stream, sess_die)
	ch_ipc := make(chan *Game_Frame, DEFAULT_CH_IPC_SIZE)

	defer func() {
		if sess.Flag&SESS_REGISTERED != 0 {
			// TODO: destroy session
			sess.Flag &^= SESS_REGISTERED
			registry.Unregister(sess.UserId)
		}
		close(sess_die)
		log.Trace("stream end:", sess.UserId)
	}()

	// >> main message loop <<
	for {
		select {
		case frame, ok := <-ch_agent: // frames from agent
			if !ok { // EOF
				return nil
			}
			switch frame.Type {
			case Game_Message:
				// validation
				if sess.Flag&SESS_REGISTERED == 0 {
					log.Critical("user not registered")
					return ERROR_USER_NOT_REGISTERED
				}

				// locate handler by proto number
				reader := packet.Reader(frame.Message)
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

				// CAUTION: serialized processing, no future locks needed.
				// multiple agents can connect simutaneously to games.
				// but possibly you will split this into smaller mutexes.
				var ret []byte
				wrap := func() { ret = handle(&sess, reader) }
				s.latch(wrap)

				// construct frame & return message from logic
				if ret != nil {
					if err := stream.Send(&Game_Frame{Type: Game_Message, Message: ret}); err != nil {
						log.Critical(err)
						return err
					}
				}

				// session control by logic
				if sess.Flag&SESS_KICKED_OUT != 0 { // logic kick out
					if err := stream.Send(&Game_Frame{Type: Game_Kick}); err != nil {
						log.Critical(err)
						return err
					}
					return nil
				}
			case Game_Register:
				if sess.Flag&SESS_REGISTERED == 0 {
					// TODO: create session
					sess.Flag |= SESS_REGISTERED
					sess.UserId = frame.UserId
					registry.Register(frame.UserId, ch_ipc)
					log.Trace("user registered")
				} else {
					log.Critical("user already registered")
				}
			case Game_Unregister:
				if sess.Flag&SESS_REGISTERED != 0 {
					// TODO: destroy session
					sess.Flag &^= SESS_REGISTERED
					registry.Unregister(sess.UserId)
					log.Trace("user unregistered")
				} else {
					log.Critical("user not registered")
				}
			case Game_Ping:
				if err := stream.Send(&Game_Frame{Type: Game_Ping, Message: frame.Message}); err != nil {
					log.Critical(err)
					return err
				}
				log.Trace("pinged")
			default:
				log.Criticalf("incorrect frame type: %v", frame.Type)
				return ERROR_INCORRECT_FRAME_TYPE
			}
		case frame := <-ch_ipc: // forward async messages from interprocess(goroutines) communication
			if err := stream.Send(frame); err != nil {
				log.Critical(err)
				return err
			}
		}
	}
}

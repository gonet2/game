package main

import (
	"errors"
	"io"
	"strconv"

	"google.golang.org/grpc/metadata"

	log "github.com/gonet2/libs/nsq-logger"
)

import (
	"game/client_handler"
	"game/misc/packet"
	. "game/proto"
	"game/registry"
	. "game/types"
)

const (
	_port = ":51000"
)

const (
	SERVICE = "[GAME]"
)

var (
	ERROR_INCORRECT_FRAME_TYPE = errors.New("incorrect frame type")
	ERROR_SERVICE_NOT_BIND     = errors.New("service not bind")
)

type server struct{}

// PIPELINE #1 stream receiver
// this function is to make the stream receiving SELECTABLE
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

// PIPELINE #2 stream processing
// the center of game logic
func (s *server) Stream(stream GameService_StreamServer) error {
	defer PrintPanicStack()
	// session init
	var sess Session
	sess_die := make(chan struct{})
	ch_agent := s.recv(stream, sess_die)
	ch_ipc := make(chan *Game_Frame, DEFAULT_CH_IPC_SIZE)

	defer func() {
		registry.Unregister(sess.UserId)
		close(sess_die)
		log.Trace("stream end:", sess.UserId)
	}()

	// read metadata from context
	md, ok := metadata.FromContext(stream.Context())
	if !ok {
		log.Critical("cannot read metadata from context")
		return ERROR_INCORRECT_FRAME_TYPE
	}
	// read key
	if len(md["userid"]) == 0 {
		log.Critical("cannot read key:userid from metadata")
		return ERROR_INCORRECT_FRAME_TYPE
	}
	// parse userid
	userid, err := strconv.Atoi(md["userid"][0])
	if err != nil {
		log.Critical(err)
		return ERROR_INCORRECT_FRAME_TYPE
	}

	// register user
	sess.UserId = int32(userid)
	registry.Register(sess.UserId, ch_ipc)
	log.Finef("userid %v logged in", sess.UserId)

	// >> main message loop <<
	for {
		select {
		case frame, ok := <-ch_agent: // frames from agent
			if !ok { // EOF
				return nil
			}
			switch frame.Type {
			case Game_Message: // the passthrough message from client->agent->game
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

				// handle request
				ret := handle(&sess, reader)

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

package client_handler

import (
	"misc/packet"
	. "types"
)

var Code = map[string]int16{}

var RCode = map[int16]string{}
var Handlers map[int16]func(*Session, *packet.Packet) []byte

func init() {
	Handlers = map[int16]func(*Session, *packet.Packet) []byte{}
}

package client_handler

import (
	"misc/packet"
	. "types"
)

var Code = map[string]int16{

	"proto_ping_req": 1001, //  ping
	"proto_ping_ack": 1002, //  ping回复
}

var RCode = map[int16]string{

	1001: "proto_ping_req", //  ping
	1002: "proto_ping_ack", //  ping回复
}
var Handlers map[int16]func(*Session, *packet.Packet) []byte

func init() {
	Handlers = map[int16]func(*Session, *packet.Packet) []byte{
		1001: P_proto_ping_req,
	}
}

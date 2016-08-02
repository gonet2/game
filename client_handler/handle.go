package client_handler

import (
	"game/misc/packet"
	. "game/types"
)

//----------------------------------- ping
func P_proto_ping_req(sess *Session, reader *packet.Packet) []byte {
	tbl, _ := PKT_auto_id(reader)
	return packet.Pack(Code["proto_ping_ack"], tbl, nil)
}

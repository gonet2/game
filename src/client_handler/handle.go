package client_handler

import (
	"misc/packet"
	. "types"

	log "github.com/gonet2/libs/nsq-logger"
)

func checkErr(err error) {
	if err != nil {
		log.Error(err)
		panic("error occured in protocol module")
	}
}

//----------------------------------- ping
func P_proto_ping_req(sess *Session, reader *packet.Packet) []byte {
	tbl, _ := PKT_auto_id(reader)
	return packet.Pack(Code["proto_ping_ack"], tbl, nil)
}

package main

import log "github.com/gonet2/libs/nsq-logger"

import (
	"misc/packet"
	. "types"
)

// 玩家登陆过程
func P_user_login_req(sess *Session, reader *packet.Packet) []byte {
	return nil
}

func checkErr(err error) {
	if err != nil {
		log.Error(err)
		panic("error occured in protocol module")
	}
}

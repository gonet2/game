package client_handler

import "misc/packet"
import . "types"

var Code = map[string]int16{
	"heart_beat_req":         0,    // 心跳包..
	"heart_beat_ack":         1,    // 心跳包回复
	"user_login_req":         10,   // 登陆
	"user_login_succeed_ack": 11,   // 登陆成功
	"user_login_faild_ack":   12,   // 登陆失败
	"client_error_ack":       13,   // 客户端错误
	"get_seed_req":           30,   // socket通信加密使用
	"get_seed_ack":           31,   // socket通信加密使用
	"proto_ping_req":         1001, //  ping
	"proto_ping_ack":         1002, //  ping回复
}

var RCode = map[int16]string{
	0:    "heart_beat_req",         // 心跳包..
	1:    "heart_beat_ack",         // 心跳包回复
	10:   "user_login_req",         // 登陆
	11:   "user_login_succeed_ack", // 登陆成功
	12:   "user_login_faild_ack",   // 登陆失败
	13:   "client_error_ack",       // 客户端错误
	30:   "get_seed_req",           // socket通信加密使用
	31:   "get_seed_ack",           // socket通信加密使用
	1001: "proto_ping_req",         //  ping
	1002: "proto_ping_ack",         //  ping回复
}

var Handlers map[int16]func(*Session, *packet.Packet) []byte

func init() {
	Handlers = map[int16]func(*Session, *packet.Packet) []byte{
		1001: P_proto_ping_req,
	}
}

package main

import "misc/packet"
import . "types"

var Code = map[string]int16{
	"user_login_req":         1, // 客户端发送用户登陆请求包
	"user_login_succeed_ack": 2, // 登陆成功
	"user_login_faild_ack":   3, // 登陆失败
	"client_error_ack":       4, // 客户端错误
}

var RCode = map[int16]string{
	1: "user_login_req",         // 客户端发送用户登陆请求包
	2: "user_login_succeed_ack", // 登陆成功
	3: "user_login_faild_ack",   // 登陆失败
	4: "client_error_ack",       // 客户端错误
}

var Handlers map[int16]func(*Session, *packet.Packet) []byte

func init() {
	Handlers = map[int16]func(*Session, *packet.Packet) []byte{
		1: P_user_login_req,
	}
}

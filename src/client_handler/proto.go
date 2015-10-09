package client_handler

import "misc/packet"

//# 该文件规定客户端和服务之间的通信结构体模式.注释必须独占一行!!!!!
//#
//# 基本类型 : integer float string boolean
//# 格式如下所示.若要定义数组，查找array看看已有定义你懂得.
//#
//# 每一个定义以'
//# 紧接一行注释 #描述这个逻辑结构用来干啥.
//# 然后定义结构名字，以'='结束，这样可以grep '=' 出全部逻辑名字.
//# 之后每一行代表一个成员定义.
//#
//# 发布代码前请确保这些部分最新.
//#
//#公共结构， 用于只传id,或一个数字的结构
type S_auto_id struct {
	F_id int32
}

func (p S_auto_id) Pack(w *packet.Packet) {
	w.WriteS32(p.F_id)

}

//#一般性回复payload,0代表成功
type S_error_info struct {
	F_code int32
	F_msg  string
}

func (p S_error_info) Pack(w *packet.Packet) {
	w.WriteS32(p.F_code)
	w.WriteString(p.F_msg)

}

//#用户登陆发包 1代表使用uuid登陆 2代表使用客户端证书登陆
type S_user_login_info struct {
	F_login_way          int32
	F_open_udid          string
	F_client_certificate string
	F_client_version     int32
	F_user_lang          string
	F_app_id             string
	F_os_version         string
	F_device_name        string
	F_device_id          string
	F_device_id_type     int32
	F_login_ip           string
}

func (p S_user_login_info) Pack(w *packet.Packet) {
	w.WriteS32(p.F_login_way)
	w.WriteString(p.F_open_udid)
	w.WriteString(p.F_client_certificate)
	w.WriteS32(p.F_client_version)
	w.WriteString(p.F_user_lang)
	w.WriteString(p.F_app_id)
	w.WriteString(p.F_os_version)
	w.WriteString(p.F_device_name)
	w.WriteString(p.F_device_id)
	w.WriteS32(p.F_device_id_type)
	w.WriteString(p.F_login_ip)

}

//#通信加密种子
type S_seed_info struct {
	F_client_send_seed    int32
	F_client_receive_seed int32
}

func (p S_seed_info) Pack(w *packet.Packet) {
	w.WriteS32(p.F_client_send_seed)
	w.WriteS32(p.F_client_receive_seed)

}

//#用户信息包
type S_user_snapshot struct {
	F_uid int32
}

func (p S_user_snapshot) Pack(w *packet.Packet) {
	w.WriteS32(p.F_uid)

}
func PKT_auto_id(reader *packet.Packet) (tbl S_auto_id, err error) {
	tbl.F_id, err = reader.ReadS32()
	checkErr(err)

	return
}

func PKT_error_info(reader *packet.Packet) (tbl S_error_info, err error) {
	tbl.F_code, err = reader.ReadS32()
	checkErr(err)

	tbl.F_msg, err = reader.ReadString()
	checkErr(err)

	return
}

func PKT_user_login_info(reader *packet.Packet) (tbl S_user_login_info, err error) {
	tbl.F_login_way, err = reader.ReadS32()
	checkErr(err)

	tbl.F_open_udid, err = reader.ReadString()
	checkErr(err)

	tbl.F_client_certificate, err = reader.ReadString()
	checkErr(err)

	tbl.F_client_version, err = reader.ReadS32()
	checkErr(err)

	tbl.F_user_lang, err = reader.ReadString()
	checkErr(err)

	tbl.F_app_id, err = reader.ReadString()
	checkErr(err)

	tbl.F_os_version, err = reader.ReadString()
	checkErr(err)

	tbl.F_device_name, err = reader.ReadString()
	checkErr(err)

	tbl.F_device_id, err = reader.ReadString()
	checkErr(err)

	tbl.F_device_id_type, err = reader.ReadS32()
	checkErr(err)

	tbl.F_login_ip, err = reader.ReadString()
	checkErr(err)

	return
}

func PKT_seed_info(reader *packet.Packet) (tbl S_seed_info, err error) {
	tbl.F_client_send_seed, err = reader.ReadS32()
	checkErr(err)

	tbl.F_client_receive_seed, err = reader.ReadS32()
	checkErr(err)

	return
}

func PKT_user_snapshot(reader *packet.Packet) (tbl S_user_snapshot, err error) {
	tbl.F_uid, err = reader.ReadS32()
	checkErr(err)

	return
}

func checkErr(err error) {
	if err != nil {
		panic("error occured in protocol module")
	}
}

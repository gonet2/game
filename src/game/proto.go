package main

import "misc/packet"

type auto_id struct {
	F_id int32
}

func (p auto_id) Pack(w *packet.Packet) {
	w.WriteS32(p.F_id)

}

type command_result_info struct {
	F_code int32
	F_msg  string
}

func (p command_result_info) Pack(w *packet.Packet) {
	w.WriteS32(p.F_code)
	w.WriteString(p.F_msg)

}

type user_login_info struct {
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

func (p user_login_info) Pack(w *packet.Packet) {
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

type pike_seed_info struct {
	F_client_send_seed    int32
	F_client_receive_seed int32
}

func (p pike_seed_info) Pack(w *packet.Packet) {
	w.WriteS32(p.F_client_send_seed)
	w.WriteS32(p.F_client_receive_seed)

}

type user_snapshot struct {
	F_uid         int32
	F_name        string
	F_level       int32
	F_current_exp int32
}

func (p user_snapshot) Pack(w *packet.Packet) {
	w.WriteS32(p.F_uid)
	w.WriteString(p.F_name)
	w.WriteS32(p.F_level)
	w.WriteS32(p.F_current_exp)

}
func PKT_auto_id(reader *packet.Packet) (tbl auto_id, err error) {
	tbl.F_id, err = reader.ReadS32()
	checkErr(err)

	return
}

func PKT_command_result_info(reader *packet.Packet) (tbl command_result_info, err error) {
	tbl.F_code, err = reader.ReadS32()
	checkErr(err)

	tbl.F_msg, err = reader.ReadString()
	checkErr(err)

	return
}

func PKT_user_login_info(reader *packet.Packet) (tbl user_login_info, err error) {
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

func PKT_pike_seed_info(reader *packet.Packet) (tbl pike_seed_info, err error) {
	tbl.F_client_send_seed, err = reader.ReadS32()
	checkErr(err)

	tbl.F_client_receive_seed, err = reader.ReadS32()
	checkErr(err)

	return
}

func PKT_user_snapshot(reader *packet.Packet) (tbl user_snapshot, err error) {
	tbl.F_uid, err = reader.ReadS32()
	checkErr(err)

	tbl.F_name, err = reader.ReadString()
	checkErr(err)

	tbl.F_level, err = reader.ReadS32()
	checkErr(err)

	tbl.F_current_exp, err = reader.ReadS32()
	checkErr(err)

	return
}

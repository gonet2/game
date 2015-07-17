package client_handler

import "misc/packet"

type auto_id struct {
	F_id int32
}

func (p auto_id) Pack(w *packet.Packet) {
	w.WriteS32(p.F_id)

}

func PKT_auto_id(reader *packet.Packet) (tbl auto_id, err error) {
	tbl.F_id, err = reader.ReadS32()
	checkErr(err)

	return
}

package types

const (
	SESS_KICKED_OUT = 0x1 // 踢掉
)

type Session struct {
	Flag   int32 // 会话标记
	UserId int32
}

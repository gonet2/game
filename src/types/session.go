package types

const (
	SESS_REGISTERED = 0x1 // 已注册
	SESS_KICKED_OUT = 0x4 // 踢掉
)

type Session struct {
	Flag   int32 // 会话标记
	UserId int32
}

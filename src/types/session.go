package types

const (
	SESS_KICKED_OUT = 0x1 // 踢掉
)

// 会话:
// 会话是一个单独玩家的上下文，在连入后到退出前的整个生命周期内存在
// 根据业务自行扩展上下文
type Session struct {
	Flag   int32 // 会话标记
	UserId int32
}

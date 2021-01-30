package raft

type LogType uint8

// 日志数据类型: 普通命令、新增节点、删除节点
const (
	LOGCOMMAND LogType = iota
	ADDSERVER
	REMOVESERVER
)

// 日志数据
type LogEntry struct {
	Index int // 索引
	Term int // 当前日志所在的任期

	Type LogType
	Date []byte // 具体的指令数据
}

// 节点状态
type State int

const(
	FOLLOWER State = iota
	CANDIDATE
	LEADER
)

type RaftNode struct {

}

package raft

import (
	"log"
	"net"
	"os"
)

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
	// 当前状态
	state State
	// endpoint
	config *RaftConfig
	// 当前Leader的地址
	leader string
	// 收到的票数
	voteCount int
	// 当前Term
	currentTerm int
	// 给谁投票
	votedFor string
	// 提交日志的最高索引
	commitIndex int
	commitTerm int
	lastApplied int

	// 发送RPC
	toCh chan int
	// 用于传输RPC
	rpcCh chan *RaftRPC
	// 用于client数据传输
	reqCh chan interface{}
	conn net.Conn
	// 日志
	Log []*LogEntry

	leaderState *LeaderState
	// 用于记录
	logger *log.Logger

}

type LeaderState struct {
	commitCh        chan interface{}
	// 存放follower节点最高log index
	replicatedIndex map[string]int
}

func newLeaderState(members []string) *LeaderState {
	ls := &LeaderState{
		commitCh:        make(chan interface{}, 1),
		replicatedIndex: make(map[string]int),
	}
	for _, member := range members {
		ls.replicatedIndex[member] = 0
	}
	return ls
}

// 初始化
func initRaft(config *RaftConfig) *RaftNode {
	ls := newLeaderState(config.members)
	r := &RaftNode{
		state:       FOLLOWER,
		config:      config,
		logger:      log.New(os.Stdout, "", log.LstdFlags),
		toCh:        make(chan int, 1),
		rpcCh:       make(chan *RaftRPC, 1),
		leaderState: ls,
		Log:         []*LogEntry{},
	}
	return r
}

// 该节点以FOLLOWER状态运行，直到状态发生变化
func (r *RaftNode) runFollower() {

}

// 该节点以CANDIDATE状态运行，直到状态发生变化
func (r *RaftNode) runCandidate() {

}

// 该节点以LEADER状态运行，直到状态发生变化
func (r *RaftNode) runLeader() {

}

// 用于处理RPC请求与响应
func (r *RaftNode) handleRPC(rpc *RaftRPC) {

}
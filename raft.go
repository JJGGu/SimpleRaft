package raft

import (
	"errors"
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

	// 用于随机选举超时接收信号
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
func (r *RaftNode) runFollower() error {
	go randomTimeout(r.toCh)
	r.logger.Printf("[INFO] %s running as follower", r.config.addr)
	for {
		select {
		case to := <- r.toCh:
			if to == TimeRunOut{
				log.Printf("时间用完了,转变为candidate")
				r.state = CANDIDATE
				r.currentTerm += 1
			}
		case rpc := <- r.rpcCh:
			r.handleRPC(rpc)
		}
	}
}

// 该节点以CANDIDATE状态运行，直到状态发生变化
func (r *RaftNode) runCandidate() {
	r.voteCount = 1
	electionTimeout := 0
	r.votedFor = r.config.addr
	majority := len(r.config.members) / 2

	for _, member := range r.config.members {
		go r.SendRequestVote(member)
	}
	// read from rpcCh until majority is achieved
	// if gets AppendEntries or VoteReq with > term revert to follower
	// else keep counting votes
	for r.state == CANDIDATE {
		go randomTimeout(r.toCh)
		select {
		case rpc := <-r.rpcCh:
			r.logger.Printf("[INFO] candidate received rpc")
			r.handleRPC(rpc)
			if r.state != CANDIDATE {
				return
			}
		case _ = <-r.toCh:
			r.logger.Printf("[INFO] %s received %d votes", r.config.addr, r.voteCount)
			electionTimeout += 1
			go randomTimeout(r.toCh)
		}
		// 转变为Leader
		if r.voteCount > majority {
			r.logger.Printf("[LEADER] taking power %s", r.config.addr)
			r.state = LEADER
			return
		}
		// 当前term未选出，重新开始选
		if electionTimeout >= 2 {
			r.currentTerm += 1
			electionTimeout = 0
			r.voteCount = 1
			r.logger.Printf("[INFO] term is %d", r.currentTerm)
			for _, member := range r.config.members {
				go r.SendRequestVote(member)
			}
		}
	}
}

// 该节点以LEADER状态运行，直到状态发生变化
func (r *RaftNode) runLeader() {
	r.logger.Printf("[INFO] %s is now leader", r.config.addr)
	go randomTimeout(r.toCh)
	for r.state == LEADER {
		// 发送心跳(存放日志信息）
		for _, member := range r.config.members {
			go r.SendHeartbeat(member)
		}
		select {
		case rpc := <-r.rpcCh:
			r.handleRPC(rpc)
		case _ = <-r.toCh:
			for _, member := range r.config.members {
				go r.SendHeartbeat(member)
			}
		}
	}
}

// 用于处理不同类型的RPC请求与响应
func (r *RaftNode) handleRPC(rpc *RaftRPC) {
	switch rpc.St.(type) {
	case AppendEntries:
		r.HandleAppendEntries(rpc.St.(AppendEntries))
	case AppendEntriesResponse:
		r.HandleAppendEntriesResponse(rpc.St.(AppendEntriesResponse))
	case RequestVote:
		r.HandleRequestVote(rpc.St.(RequestVote))
	case RequestVoteResponse:
		r.HandleRequestVoteResponse(rpc.St.(RequestVoteResponse), &r.voteCount)
	case ClientRequest:
		r.HandleClientRequest(rpc.St.(ClientRequest))
	default:
		r.logger.Printf("[ERROR] 未知 RPC: %+v\n", rpc)
	}
}

func (r *RaftNode) RunRaft() error {
	// 监听端口
	go r.RunTCPServer()

	for {
		if r.state == FOLLOWER {
			r.runFollower()
		} else if r.state == CANDIDATE {
			r.runCandidate()
		} else if r.state == LEADER {
			r.runLeader()
		} else {
			return errors.New("state error")
		}

	}
}
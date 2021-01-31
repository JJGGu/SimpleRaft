package raft

// 用于RPC请求与相应

type RaftRPC struct {
	St interface{}
}
// 请求投票
type RequestVote struct {
	CandidateTerm int
	CandidateId string // 候选者地址
	LastLogIndex int // 该候选者最后一个Log索引
	LastLogTerm int // 该候选者最后一个Log的term
}

// 投票响应
type RequestVoteResponse struct {
	VoteId string // 投票者的地址
	VoteGranted bool // 当投票者所在的任期大于候选者的时候设为false
}

// 用于日志一致性同步和心跳
type AppendEntries struct {
	LeaderTerm int // 当前任期
	Addr string // Leader地址

	LeaderCommit int // 该Leader提交的最高的Log index
	PrevLogIndex int // 这个新的追加日志前的一个（当前最新的）Log索引
	PrevLogTerm int // PrevLogIndex 对应的Term

	NewEntries []*LogEntry // 日志数据,可能有多条日志
}

type AppendEntriesResponse struct {
	Success bool // 当follower的Term大于leader时，或者Term相同但是LogIndex大于Leader的LogIndex时设置为false
	Term int // Follower 的Term
	FollowerCommit int // Follower当前已提交的最大的LogIndex， Leader根据此设置对该Follower的nextIndex
	Addr string
}

// 客户端的请求（简化版）
type ClientRequest struct {
	Entry *LogEntry
}

// 下面是具体的RPC请求与处理的方法

// 处理从客户端传来的请求
func (r *RaftNode) HandleClientRequest(request ClientRequest)  {
	// 如果是发送给FOLLOWER或者CANDIDATE则将请求重定向到Leader
	if r.state == FOLLOWER || r.state == CANDIDATE {
		go SendStruct(r.leader, request)
	} else if r.state == LEADER {
		// 日志追加
		r.Log = append(r.Log, request.Entry)
		// 简化处理
		r.commitIndex += 1
	}
}

// 心跳&logEntries
func (r *RaftNode) SendHeartbeat(addr string) error {
	prevLogIndex := r.leaderState.replicatedIndex[addr]
	prevLogTerm := r.Log[prevLogIndex].Term
	ae := AppendEntries{
		// leader metadata
		LeaderTerm:   r.currentTerm,
		Addr:         r.config.addr,
		LeaderCommit: r.commitIndex,

		// log info
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		NewEntries:   r.Log[prevLogIndex:],
	}
	go SendStruct(addr, ae)
	return nil
}

// 向某个节点发送投票RPC请求
func (r *RaftNode) SendRequestVote(addr string) error {
	rv := RequestVote{
		CandidateTerm: r.currentTerm,
		CandidateId:   r.config.addr,

		LastLogIndex: r.commitIndex,
		LastLogTerm:  r.commitTerm,
	}
	go SendStruct(addr, rv)
	return nil
}

// 处理收到的投票请求响应信息
func (r *RaftNode) HandleRequestVoteResponse(response RequestVoteResponse, count *int) {
	// 只有CANDIDATE需要处理
	if r.state == CANDIDATE && response.VoteGranted {
		*count += 1
		r.logger.Printf("收到投票")
	}
}

// 收到请求投票RPC后进行处理
func (r *RaftNode) HandleRequestVote(request RequestVote) error {
	// 返回信息
	response := RequestVoteResponse{}
	if r.state == FOLLOWER {
		// 如果是FOLLOWER收到，那就进行判断（这里只判断了term，论文中还需要判断Logindex），然后决定是否投票
		if r.currentTerm < request.CandidateTerm {
			response.VoteGranted = true
			r.votedFor = request.CandidateId
		}
	} else if r.state == CANDIDATE || r.state == LEADER {
		// 判断状态是否需要变化
		if r.currentTerm < request.CandidateTerm {
			r.state = FOLLOWER
			response.VoteGranted = true
			r.votedFor = request.CandidateId
		}
	}
	if response.VoteGranted == true {
		r.logger.Printf("[INFO] voting true")
	}
	go SendStruct(request.CandidateId, response)
	return nil
}

// 收到从heartbeat传来的日志后，进行处理
func (r *RaftNode) HandleAppendEntries(request AppendEntries) error {
	// 如果当前节点是Candidate或者Leader，说明有其他节点是Leader
	if r.state == CANDIDATE || r.state == LEADER {
		if r.currentTerm <= request.LeaderTerm {
			// =是有可能在同一任期，选举的时候
			r.state = FOLLOWER
		}
	}
	response := AppendEntriesResponse{
		Addr: r.config.addr,
	}
	if r.currentTerm > request.LeaderTerm {
		response.Success = false
		go SendStruct(request.Addr, response)
		return nil
	} else {
		response.Success = true
	}

	// TODO LogIndex 检查, follower可能由于宕机部分日志没有同步
	// 更新当前节点的Log
	r.Log = append(r.Log[:request.PrevLogIndex], request.NewEntries...)
	r.currentTerm = request.LeaderTerm
	r.commitIndex = len(r.Log) - 1

	response.Term = r.currentTerm
	response.FollowerCommit = r.commitIndex
	go SendStruct(request.Addr, response)
	return nil
}

// 用于Leader处理收到其他节点返回的响应信息
func (r *RaftNode) HandleAppendEntriesResponse(response AppendEntriesResponse) error {
	// 追加日志成功
	if response.Success {
		r.leaderState.replicatedIndex[response.Addr] = response.FollowerCommit
	} else {
		r.state = FOLLOWER
	}
	return nil
}



package raft

// 用于RPC请求与相应

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

}

func (r *RaftNode) SendRequestVote(addr string) error {

}

func (r *RaftNode) HandleRequestVote(request RequestVote)  {

}

func (r *RaftNode) HandleRequestVoteResponse(response RequestVoteResponse, count *int)  {

}

func (r *RaftNode) SendHeartbeat(addr string) error {

}

func (r *RaftNode) HandleAppendEntries(request AppendEntries) error {


}

func (r *RaftNode) HandleAppendEntriesResponse(response AppendEntriesResponse) error {

}



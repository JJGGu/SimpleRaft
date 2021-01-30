package raft

type RaftConfig struct {
	addr string // 当前节点的地址
	members []string // 所有节点的地址
}

func DefaultConfig() *RaftConfig  {
	config := &RaftConfig{
		addr: "localhost:7999",
	}

	return config
}

func CreateConfig(addr string, members []string) *RaftConfig {
	config := &RaftConfig{
		addr: addr,
		members: members,
	}
	return config
}
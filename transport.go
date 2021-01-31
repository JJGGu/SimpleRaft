package raft

import (
	"encoding/gob"
	"log"
	"net"
)

// 底层数据传输
func HandleConnection(conn net.Conn, ch chan *RaftRPC) {
	dec := gob.NewDecoder(conn)
	msg := &RaftRPC{}
	dec.Decode(msg)
	ch <- msg
}

// RPC通信,监听端口
func (r *RaftNode) RunTCPServer() {
	lis, err := net.Listen("tcp", ":" + getPort(r.config.addr))
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go HandleConnection(conn, r.rpcCh)
	}
}

// 发送数据
func SendStruct(addr string, st interface{}) {
	conn,err := net.Dial("tcp", addr)
	if err != nil {
		return
	}
	rrpc := &RaftRPC{
		St: st,
	}
	gob.Register(rrpc.St)
	encoder := gob.NewEncoder(conn)
	err = encoder.Encode(rrpc)
	if err != nil{
		return
	}
	conn.Close()
}
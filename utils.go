package raft

import (
	"math/rand"
	"strings"
	"time"
)

// 随机选举超时时间50-300ms
const TimeoutMax = 300
const TimeoutMin = 50
const TimeRunOut = 0

func getPort(addr string) string {
	parts := strings.Split(addr, ":")
	return parts[len(parts) - 1]
}

func random(min, max int) int {
	return rand.Intn(max - min) + min
}

// 时间一到就输入到tochan
func randomTimeout(toChan chan int) {
	randomTime := time.Duration(random(TimeoutMin, TimeoutMax))
	time.Sleep(randomTime * time.Millisecond)
	toChan <- TimeRunOut
}


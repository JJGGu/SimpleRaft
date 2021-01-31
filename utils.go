package raft

import "strings"

func getPort(addr string) string {
	parts := strings.Split(addr, ":")
	return parts[len(parts) - 1]
}

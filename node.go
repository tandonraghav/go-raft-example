package main

import (
	"github.com/hashicorp/raft"
)

type Node struct {
	RaftDir  string
	RaftBind string
	raft     *raft.Raft
	fsm      *fsm
}

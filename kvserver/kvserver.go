// Copyright (c) 2023 yixiak
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package kvserver

import (
	"raftkv/kvnode"
	"sync"
)

type KVserver struct {
	mu        sync.Mutex
	node      *kvnode.KVnode
	me        int32
	storage   map[string]int
	applychan chan kvnode.ApplyMsg
}

func NewKVServer(me int, peers []int, addrs []string, persist string) *KVserver {

	applych := make(chan kvnode.ApplyMsg)
	inner := kvnode.NewKVnode(me, peers, addrs, applych)

	kvServer := &KVserver{
		me:        int32(me),
		node:      inner,
		applychan: applych,
	}
	return kvServer
}

func (server *KVserver) Connect() {
	server.node.Connect()
}

func (server *KVserver) IsLeader() bool {
	return server.node.IsLeader()
}

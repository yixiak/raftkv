// Copyright (c) 2023 yixiak
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package kvserver

import (
	"fmt"
	"raftkv/kvnode"
	"sync"
)

type KVserver struct {
	mu        sync.Mutex
	node      *kvnode.KVnode
	me        int32
	storage   map[string]int
	applychan chan kvnode.ApplyMsg
	chanmap   map[int]chan OpMsg
	killed    bool
}

type OpMsg struct {
	Index int32
	Succ  bool
	Op    string
	Key   string
	Value int32
	// fail massage
	Msg string
}

func NewKVServer(me int, peers []int, addrs []string, persist string) *KVserver {

	applych := make(chan kvnode.ApplyMsg)
	inner := kvnode.NewKVnode(me, peers, addrs, applych)

	kvServer := &KVserver{
		me:        int32(me),
		node:      inner,
		applychan: applych,
		killed:    false,
	}
	return kvServer
}

func (server *KVserver) Connect() {
	server.node.Connect()
	go server.ticker()
}

func (server *KVserver) IsLeader() bool {
	return server.node.IsLeader()
}

func (kv *KVserver) Exec(ch chan OpMsg, op string, key string, value int32) {

	index, term, isleader := kv.node.Exec(op, key, value)
	if index != -1 && term != -1 && isleader {
		kv.mu.Lock()
		kv.chanmap[index] = ch
		kv.mu.Unlock()
	}
}

func (kv *KVserver) ticker() {
	for !kv.killed {
		// get an apply msg
		msg := <-kv.applychan
		opmsg := kv.apply(&msg)
		index := msg.Index
		kv.chanmap[index] <- opmsg
	}
}

func (kv *KVserver) apply(msg *kvnode.ApplyMsg) OpMsg {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	opmsg := &OpMsg{
		Index: int32(msg.Index),
		Succ:  true,
		Op:    msg.Op,
		Key:   msg.Key,
		Value: msg.Value,
	}
	switch msg.Op {
	case "Put", "Update":
		kv.storage[msg.Key] = int(msg.Value)
	case "remove":
		_, succ := kv.storage[msg.Key]
		if succ {
			delete(kv.storage, msg.Key)
		} else {
			opmsg.Succ = false
			opmsg.Msg = fmt.Sprintf("delete fail: there is no %v", msg.Key)
		}
	case "find":
		value, succ := kv.storage[msg.Key]
		if succ {
			opmsg.Value = int32(value)
		} else {
			opmsg.Succ = false
			opmsg.Msg = fmt.Sprintf("find fail: there is no %v", msg.Key)
		}
	}
	return *opmsg
}

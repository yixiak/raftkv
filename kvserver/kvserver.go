package kvserver

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"raftkv/debug"
	"raftkv/kvnode"
	"sync"

	"google.golang.org/grpc"
)

type KVserver struct {
	mu         sync.Mutex
	node       *kvnode.KVnode
	me         int32
	storage    map[string]int
	applychan  chan kvnode.ApplyMsg
	chanmap    map[int]chan OpMsg
	lis        net.Listener
	grpcServer *grpc.Server
	filePath   string
	killed     bool
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
	debug.Dlog("[Server %v] is making node", me)
	inner := kvnode.NewKVnode(me, peers, addrs, applych)
	chmap := make(map[int]chan OpMsg)
	storage := make(map[string]int)

	//go kvServer.ticker()

	lis, err := net.Listen("tcp", addrs[me])
	if err != nil {
		panic("listen failed")
	}
	grpcSever := grpc.NewServer()
	kvnode.RegisterRaftKVServer(grpcSever, inner)
	// should run at another thread
	go grpcSever.Serve(lis)
	kvServer := &KVserver{
		me:         int32(me),
		node:       inner,
		applychan:  applych,
		killed:     false,
		chanmap:    chmap,
		storage:    storage,
		lis:        lis,
		grpcServer: grpcSever,
		filePath:   persist,
	}
	return kvServer
}

func (server *KVserver) Connect() {
	debug.Dlog("[Server ] is connecting with others")
	server.node.Connect()
	go server.ticker()
}

func (server *KVserver) IsLeader() bool {
	return server.node.IsLeader()
}

func (kv *KVserver) Exec(ch chan OpMsg, op string, key string, value int32) {
	debug.Dlog("[Server %v] receive a op request : %v %v", kv.me, op, key)

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
		//debug.Dlog("[Server %v] is blocking for apply message from node", kv.me)
		msg := <-kv.applychan
		//debug.Dlog("[Server %v] receive apply message:%v from node", kv.me, msg.Index)
		opmsg := kv.apply(&msg)
		index := msg.Index
		ch := kv.getReplychan(index)
		if ch != nil {
			debug.Dlog("[Server %v] is sending msg to DBSevice", kv.me)
			ch <- opmsg
			debug.Dlog("[Server %v] finished sending msg to DBSevice", kv.me)
		}
	}
}

// Only one server should send msg to Service
// the other
func (kv *KVserver) getReplychan(index int) chan OpMsg {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	if _, ok := kv.chanmap[index]; !ok {
		// if the server doesn't need to send msg
		// this chan will be cover.
		return nil
	}
	return kv.chanmap[index]
}

func (kv *KVserver) apply(msg *kvnode.ApplyMsg) OpMsg {
	debug.Dlog("[Server %v] is applying %v", kv.me, msg.Index)
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
	case "put", "update":
		kv.storage[msg.Key] = int(msg.Value)
		_, succ := kv.storage[msg.Key]
		if succ {
			debug.Dlog("[Server %v] %v the %v:%v successfully", kv.me, msg.Op, msg.Key, msg.Value)
		}
	case "remove":
		_, succ := kv.storage[msg.Key]
		if succ {
			delete(kv.storage, msg.Key)
			debug.Dlog("[Server %v] delete %v successfully", kv.me, msg.Key)
		} else {
			opmsg.Succ = false
			opmsg.Msg = fmt.Sprintf("delete fail: there is no %v", msg.Key)
		}
	case "get":
		value, succ := kv.storage[msg.Key]
		if succ {
			opmsg.Value = int32(value)
			debug.Dlog("[Server %v] get %v successfully", kv.me, msg.Key)
		} else {
			opmsg.Succ = false
			opmsg.Msg = fmt.Sprintf("get fail: there is no %v", msg.Key)
		}
	}
	debug.Dlog("[Server %v] finish apply", kv.me)
	return *opmsg
}

func (kv *KVserver) Close() error {
	kv.mu.Lock()
	kv.killed = true
	kv.mu.Unlock()
	err := kv.node.Close()
	if err != nil {
		panic(err)
	}
	kv.grpcServer.Stop()
	kv.lis.Close()
	kv.persist()
	return nil
}

func (kv *KVserver) persist() error {
	bytes, err := json.Marshal(kv.storage)
	if err != nil {
		panic(err)
	}
	debug.Dlog("[Server %v] 's storage is %+v", kv.me, kv.storage)
	file, err := os.OpenFile(kv.filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	file.Write(bytes)
	file.Close()
	return nil
}

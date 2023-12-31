package kvnode

import (
	"net"
	"raftkv/debug"
	"testing"
	"time"

	"google.golang.org/grpc"
)

func TestMake(t *testing.T) {
	addrs := make([]string, 3)
	peers := make([]int, 3)
	addrs[0] = "127.0.0.1:10001"
	addrs[1] = "127.0.0.1:10002"
	addrs[2] = "127.0.0.1:10003"
	peers[0] = 0
	peers[0] = 1
	peers[0] = 2
	//persist := "./temp"

	grpcnode := make([]*grpc.Server, 3)
	servers := make([]*KVnode, 3)
	debug.Dlog("Testing Make:")
	// should I make listeners here?
	for index := 0; index < 3; index++ {
		server := NewKVnode(index, peers, addrs, nil)
		lis, err := net.Listen("tcp", addrs[index])
		if err != nil {
			panic("listen failed")
		}
		grpcSever := grpc.NewServer()
		RegisterRaftKVServer(grpcSever, server)
		// should run at another thread
		go grpcSever.Serve(lis)
		grpcnode[index] = grpcSever
		servers[index] = server

	}
	for index := 0; index < 3; index++ {
		servers[index].Connect()
	}
	time.Sleep(500 * time.Millisecond)
}

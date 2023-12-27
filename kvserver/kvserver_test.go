// Copyright (c) 2023 yixiak
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package kvserver

import (
	"fmt"
	"testing"
	"time"
)

func TestCreateServer(t *testing.T) {
	addrs := make([]string, 3)
	peers := make([]int, 3)
	addrs[0] = "127.0.0.1:10001"
	addrs[1] = "127.0.0.1:10002"
	addrs[2] = "127.0.0.1:10003"
	peers[0] = 0
	peers[0] = 1
	peers[0] = 2
	fmt.Printf("Test Create Server\n")
	servers := make([]*KVserver, 3)
	for i := 0; i < 3; i++ {
		server := NewKVServer(i, peers, addrs, "./temp")
		servers[i] = server

	}
	for i := 0; i < 3; i++ {
		servers[i].Connect()
	}
	time.Sleep(10 * time.Second)

}

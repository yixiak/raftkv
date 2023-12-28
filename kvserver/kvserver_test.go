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

func TestSendCommand(t *testing.T) {
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
	time.Sleep(50 * time.Millisecond)
	msgch := make(chan OpMsg)
	op := "put"
	key := "a"
	value := int32(1)
	for {
		if servers[0].IsLeader() {
			fmt.Printf("Server 0 is Leader\n")
			servers[0].Exec(msgch, op, key, value)
			break
		} else if servers[1].IsLeader() {
			fmt.Printf("Server 1 is Leader\n")
			servers[1].Exec(msgch, op, key, value)
			break
		} else if servers[2].IsLeader() {
			fmt.Printf("Server 2 is Leader\n")
			servers[2].Exec(msgch, op, key, value)
			break
		}
	}
	outtime := time.NewTimer(1 * time.Second)
	select {
	case msg := <-msgch:
		if !msg.Succ {
			t.Fatalf("Fail to apply")
		}
		fmt.Printf("1. receive op msg:%+v\n", msg)
	case <-outtime.C:
		t.Fatalf("Exec out of time")
	}
	for {
		if servers[0].IsLeader() {
			fmt.Printf("Server 0 is Leader\n")
			servers[0].Exec(msgch, op, key, value)
			break
		} else if servers[1].IsLeader() {
			fmt.Printf("Server 1 is Leader\n")
			servers[1].Exec(msgch, op, key, value)
			break
		} else if servers[2].IsLeader() {
			fmt.Printf("Server 2 is Leader\n")
			servers[2].Exec(msgch, op, key, value)
			break
		}
	}
	outtime.Reset(1 * time.Second)
	select {
	case msg := <-msgch:
		if !msg.Succ {
			t.Fatalf("Fail to apply")
		}
		fmt.Printf("2. receive op msg:%+v\n", msg)
	case <-outtime.C:
		t.Fatalf("Exec out of time")
	}

}

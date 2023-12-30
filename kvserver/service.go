package kvserver

import (
	"fmt"
	"sync"
	"time"
)

type DBService struct {
	mu        sync.Mutex
	clientnum int
	servers   []*KVserver
	closed    bool
}

func Open() *DBService {
	// use 3 servers to test
	addrs := make([]string, 3)
	peers := make([]int, 3)
	addrs[0] = "127.0.0.1:10001"
	addrs[1] = "127.0.0.1:10002"
	addrs[2] = "127.0.0.1:10003"
	peers[0] = 0
	peers[1] = 1
	peers[2] = 2

	servers := make([]*KVserver, 3)
	for i := 0; i < 3; i++ {
		server := NewKVServer(i, peers, addrs, "./temp")
		servers[i] = server
	}
	DB := &DBService{
		clientnum: 3,
		servers:   servers,
		closed:    false,
	}
	for i := 0; i < 3; i++ {
		servers[i].Connect()
	}
	return DB
}

func (db *DBService) Close() error {
	// Wait for other operations to complete
	for {
		db.mu.Lock()
		if db.clientnum == 0 {
			db.mu.Unlock()
			break
		}
		db.mu.Unlock()
		time.Sleep(1 * time.Millisecond)
	}
	db.mu.Lock()
	db.closed = true
	db.mu.Unlock()
	for i := range db.servers {
		err := db.servers[i].Close()
		if err != nil {
			return err
		}
	}
	fmt.Println("DB service has been closed()")
	return nil
}

func (db *DBService) Put(key string, value int) error {
	channel := make(chan OpMsg)
	for i := range db.servers {
		if db.servers[i].IsLeader() {
			db.mu.Lock()
			db.clientnum++
			db.mu.Unlock()
			db.servers[i].Exec(channel, "put", key, int32(value))
			break
		}
	}
	msg := <-channel
	db.mu.Lock()
	db.clientnum--
	db.mu.Unlock()
	fmt.Println(msg.Msg)
	return nil
}

func (db *DBService) Update(key string, value int) error {
	channel := make(chan OpMsg)
	for i := range db.servers {
		if db.servers[i].IsLeader() {
			db.mu.Lock()
			db.clientnum++
			db.mu.Unlock()
			db.servers[i].Exec(channel, "update", key, int32(value))
			break
		}
	}
	msg := <-channel
	db.mu.Lock()
	db.clientnum--
	db.mu.Unlock()
	fmt.Println(msg.Msg)
	return nil
}

func (db *DBService) Remove(key string) error {
	channel := make(chan OpMsg)
	for i := range db.servers {
		if db.servers[i].IsLeader() {
			db.mu.Lock()
			db.clientnum++
			db.mu.Unlock()
			db.servers[i].Exec(channel, "remove", key, 0)
			break
		}
	}
	msg := <-channel
	db.mu.Lock()
	db.clientnum--
	db.mu.Unlock()
	fmt.Println(msg.Msg)
	return nil
}

func (db *DBService) Get(key string) (int, error) {
	channel := make(chan OpMsg)
	for i := range db.servers {
		if db.servers[i].IsLeader() {
			db.mu.Lock()
			db.clientnum++
			db.mu.Unlock()
			db.servers[i].Exec(channel, "get", key, 0)
			break
		}
	}
	msg := <-channel
	db.mu.Lock()
	db.clientnum--
	db.mu.Unlock()
	fmt.Println(msg.Msg)
	return int(msg.Value), nil
}

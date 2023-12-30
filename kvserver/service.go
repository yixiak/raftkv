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

func (db *DBService) Open() bool {

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

func (db *DBService) Find(key string) (int, error) {
	channel := make(chan OpMsg)
	for i := range db.servers {
		if db.servers[i].IsLeader() {
			db.mu.Lock()
			db.clientnum++
			db.mu.Unlock()
			db.servers[i].Exec(channel, "find", key, 0)
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

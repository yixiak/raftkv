package kvserver

import (
	"fmt"
	"os"
	"raftkv/debug"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type DBService struct {
	mu        sync.Mutex
	clientnum int
	servers   []*KVserver
	closed    bool
}

type Config struct {
	Mode          string   `yaml:"mode"`
	ServersNum    int      `yaml:"serversNum"`
	Address       []string `yaml:"address"`
	Port          []int    `yaml:"port"`
	PersistPath   []string `yaml:"persistPath"`
	ListenAddress string   `yaml:"listenAddress"`
	ListenPort    string   `yaml:"listenPort"`
}

func Open() *DBService {
	var config Config
	configdata, err := os.ReadFile("./config.yaml")
	if err != nil {
		config.Mode = "default"
	}

	err = yaml.Unmarshal(configdata, &config)
	if err != nil {
		panic(err)
	}
	num := 0
	if config.Mode == "default" {
		// use 3 servers to test
		num = 3
	} else {
		num = config.ServersNum
	}

	addrs := make([]string, num)
	peers := make([]int, num)
	servers := make([]*KVserver, num)
	persists := make([]string, num)
	if config.Mode == "default" {
		addrs[0] = "127.0.0.1:10001"
		addrs[1] = "127.0.0.1:10002"
		addrs[2] = "127.0.0.1:10003"
		peers[0] = 0
		peers[1] = 1
		peers[2] = 2
		persists[0] = "./temp/temp1.json"
		persists[1] = "./temp/temp2.json"
		persists[2] = "./temp/temp3.json"
	} else {
		for i := 0; i < num; i++ {
			peers[i] = i
			addrs[i] = fmt.Sprintf("%v:%v", config.Address[i], config.Port[i])
			persists[i] = config.PersistPath[i]
		}
	}

	for i := 0; i < num; i++ {
		server := NewKVServer(i, peers, addrs, persists[i])
		servers[i] = server
	}
	DB := &DBService{
		clientnum: 0,
		servers:   servers,
		closed:    false,
	}
	for i := 0; i < num; i++ {
		servers[i].Connect()
	}
	return DB
}

func (db *DBService) Close() error {
	// avoid calling the close too quickly
	time.Sleep(5 * time.Second)
	// Wait for other operations to complete
	fmt.Println("DBService is closing")
	for {
		db.mu.Lock()
		if db.clientnum == 0 {
			debug.Dlog("[DBService] can Close now")
			db.closed = true
			db.mu.Unlock()
			break
		}
		db.mu.Unlock()
	}
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
	debug.Dlog("[DBService] receive a put request")
	channel := make(chan OpMsg)
	find := false
	for {
		if find {
			break
		}
		for i := range db.servers {
			if db.servers[i].IsLeader() {
				debug.Dlog("[DBService] find leader %v", i)
				db.mu.Lock()
				db.clientnum++
				debug.Dlog("[DBService] clientnum is %v", db.clientnum)
				db.mu.Unlock()
				db.servers[i].Exec(channel, "put", key, int32(value))
				find = true
				break
			}
		}
	}
	msg := <-channel
	db.mu.Lock()
	db.clientnum--
	debug.Dlog("[DBService] clientnum is %v", db.clientnum)
	db.mu.Unlock()
	if msg.Succ {
		fmt.Println("DBService finish put operation")
	} else {
		fmt.Println(msg.Msg)
	}
	return nil
}

func (db *DBService) Update(key string, value int) error {
	channel := make(chan OpMsg)
	find := false
	for {
		if find {
			break
		}
		for i := range db.servers {
			if db.servers[i].IsLeader() {
				db.mu.Lock()
				db.clientnum++
				debug.Dlog("[DBService] clientnum is %v", db.clientnum)
				db.mu.Unlock()
				db.servers[i].Exec(channel, "update", key, int32(value))
				find = true
				break
			}
		}
	}
	msg := <-channel
	db.mu.Lock()
	db.clientnum--
	debug.Dlog("[DBService] clientnum is %v", db.clientnum)
	db.mu.Unlock()
	if !msg.Succ {
		fmt.Println(msg.Msg)
	} else {
		fmt.Println("DBService finish update operation")
	}
	return nil
}

func (db *DBService) Remove(key string) error {
	channel := make(chan OpMsg)
	find := false
	for {
		if find {
			break
		}
		for i := range db.servers {
			if db.servers[i].IsLeader() {
				db.mu.Lock()
				db.clientnum++
				debug.Dlog("[DBService] clientnum is %v", db.clientnum)
				db.mu.Unlock()
				db.servers[i].Exec(channel, "remove", key, 0)
				find = true
				break
			}
		}
	}
	msg := <-channel
	db.mu.Lock()
	db.clientnum--
	debug.Dlog("[DBService] clientnum is %v", db.clientnum)
	db.mu.Unlock()
	if !msg.Succ {
		fmt.Println(msg.Msg)
	} else {
		fmt.Println("DBService finish remove operation")
	}
	return nil
}

func (db *DBService) Get(key string) (int, error) {
	channel := make(chan OpMsg)
	find := false
	for {
		if find {
			break
		}
		for i := range db.servers {
			if db.servers[i].IsLeader() {
				db.mu.Lock()
				db.clientnum++
				debug.Dlog("[DBService] clientnum is %v", db.clientnum)
				db.mu.Unlock()
				db.servers[i].Exec(channel, "get", key, 0)
				find = true
				break
			}
		}
	}
	msg := <-channel
	db.mu.Lock()
	db.clientnum--
	debug.Dlog("[DBService] clientnum is %v", db.clientnum)
	db.mu.Unlock()
	if !msg.Succ {
		fmt.Println(msg.Msg)
	} else {
		fmt.Printf("DBService get the value of %v:%v\n", key, msg.Value)
	}
	return int(msg.Value), nil
}

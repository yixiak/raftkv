package kv

import (
	"context"
	"math/rand"
	"raftkv/debug"
	"sync"
	"time"
)

type State int

const (
	Follower  State = 1
	Candidate State = 1
	Leader    State = 2
)

type KVserver struct {
	UnimplementedRaftKVServer

	mu sync.Mutex
	me int32

	// peers record the index of client
	// stubs is used to Calling service methods
	peers []int
	stubs []*raftKVClient

	state State

	currentTerm int32
	votedFor    int32
	logs        []*LogEntry

	commitIndex int32
	lastApplied int32

	nextIndex  []int32
	matchIndex []int32

	election_timeout  *time.Timer
	heartbeat_timeout *time.Timer

	// kv storage in memory
	storage map[string]int32
	// Persistence path
	persist string
}

// rpc server's interface
func (rf *KVserver) AppendEntries(ctx context.Context, args *AppendEntriesArgs) (*AppendEntriesReply, error) {
	reply := AppendEntriesReply{}

	return &reply, nil
}

func (rf *KVserver) RequestVote(ctx context.Context, args *RequestVoteArgs) (*RequestVoteReply, error) {
	reply := RequestVoteReply{}

	return &reply, nil
}

func (rf *KVserver) Operate(ctx context.Context, args *Operation) (*Opreturn, error) {
	reply := Opreturn{}

	return &reply, nil
}

// communicate with other servers
func (rf *KVserver) SendAppendEntries(server int, args *AppendEntriesArgs) (*AppendEntriesReply, bool) {
	return nil, true
}
func (rf *KVserver) SendRequestVote(server int, args *RequestVoteArgs) (*RequestVoteReply, bool) {
	return nil, true
}

// used to create a new server for RegisterRaftKVServer()
func newKVServer(me int, peers []int, addrs []string, persist string) *KVserver {
	rf := KVserver{}

	go rf.ticker()
	return &rf
}

// The ticker go routine starts a new election if this peer hasn't received
// heartsbeats recently, or send a heartbeat periodically if it is a Leader
func (rf *KVserver) ticker() {}

func (rf *KVserver) TobeLeader() {
	rf.election_timeout.Reset(RandElectionTimeout())
	if len(rf.logs) > 0 {
		debug.Dlog("[Server %v] to be a leader with logs len: %v , term: %v and commitIndex: %v", rf.me, len(rf.logs), rf.currentTerm, rf.commitIndex)
	}
	rf.state = Leader
	lastIndex := len(rf.logs)
	for peer := range rf.nextIndex {
		rf.nextIndex[peer] = int32(lastIndex)
		rf.matchIndex[peer] = 0
	}

	//rf.SendheartbeatToAll()
	rf.heartbeat_timeout.Reset(20 * time.Millisecond)
}

// generate a rand election time
func RandElectionTimeout() time.Duration {
	source := rand.NewSource(time.Now().UnixMicro())
	ran := rand.New(source)
	return time.Duration(500+ran.Int()%150) * time.Millisecond
}

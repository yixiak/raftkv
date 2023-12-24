package raftkv

import (
	"context"
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

	mu    sync.Mutex
	me    int32
	peers []int

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

	storage map[string]int32
}

func (rf *KVserver) AppendEntries(ctx context.Context, args *AppendEntriesArgs) (*AppendEntriesReply, error) {
	reply := AppendEntriesReply{}

	return &reply, nil
}

func (rf *KVserver) RequestVote(ctx context.Context, args *RequestVoteArgs) (*RequestVoteReply, error) {
	reply := RequestVoteReply{}

	return &reply, nil
}

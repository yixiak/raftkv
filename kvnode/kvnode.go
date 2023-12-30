package kvnode

import (
	"context"
	"fmt"
	"math/rand"
	"raftkv/debug"
	"sync"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type State int

const (
	Follower  State = 1
	Candidate State = 1
	Leader    State = 2
)

type KVnode struct {
	UnimplementedRaftKVServer

	mu sync.Mutex
	me int32

	// peers record the index of client
	// conns is used to manage  connections
	// stubs is used to Calling service methods
	peers []int
	conns []*grpc.ClientConn
	stubs []*RaftKVClient
	addrs []string

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

	// // kv storage in memory
	// storage map[string]int32
	// // Persistence path
	// persist string
	applych chan ApplyMsg

	killed bool
}

type ApplyMsg struct {
	Op    string
	Key   string
	Value int32
	Index int
}

// rpc server's interface
func (rf *KVnode) AppendEntries(ctx context.Context, args *AppendEntriesArgs) (*AppendEntriesReply, error) {
	reply := &AppendEntriesReply{}
	reply.Term = rf.currentTerm
	reply.Success = false
	if args.GetTerm() < reply.Term {
		return reply, nil
	}
	if int(args.GetPrevlogIndex()) > len(rf.logs) {
		debug.Dlog("[Node %v] return false to %v's AppendEntries for arg.prevlogindex %v > len(log)", rf.me, args.GetLeaderId(), args.GetPrevlogIndex())
		return reply, nil
	}
	rf.mu.Lock()
	defer rf.mu.Unlock()
	// Receiver implementation
	prevLogIndex := args.GetPrevlogIndex()
	if int(prevLogIndex) >= len(rf.logs) {
		return reply, nil
	}
	if args.GetPrevlogIndex() >= 0 && rf.logs[args.GetPrevlogIndex()].Term != args.GetPrevLogTerm() {
		debug.Dlog("[Node %v] Leader's term is different", rf.me)
		return reply, nil
	}
	if len(args.Entries) == 0 {
		// receive a heartbeat
		debug.Dlog("[Node %v] receive a empty Entry from %v", rf.me, args.GetLeaderId())
		rf.election_timeout.Reset(RandElectionTimeout())
		reply.Success = true
		// Leader update its commitIndex after commit itself
		if args.GetLeaderCommit() > rf.commitIndex {
			debug.Dlog("[Node %v] receive a LEADERCOMMIT %v ", rf.me, args.GetLeaderCommit())
			if int(args.GetLeaderCommit()) > len(rf.logs)-1 {
				rf.commitIndex = int32(len(rf.logs) - 1)
			} else {
				rf.commitIndex = args.GetLeaderCommit()
			}
			rf.apply()
			debug.Dlog("[Node %v] update its commitIndex to %v ", rf.me, rf.commitIndex)
		}
		return reply, nil
	}
	rf.election_timeout.Reset(RandElectionTimeout())
	debug.Dlog("[Node %v] receive appendentries with entry %+v. And currentTerm is %v, logs len is %v, committedIndex is: %v", rf.me, args.Entries[0], rf.currentTerm, len(rf.logs), rf.commitIndex)

	rf.logs = rf.logs[:prevLogIndex+1]
	rf.logs = append(rf.logs, args.Entries...)
	//debug.Dlog("[Node %v]'s loglen is %v after append", rf.me, len(rf.logs))
	if args.GetLeaderCommit() > rf.commitIndex {
		debug.Dlog("[Node %v]'s commit Index is less then Leader's", rf.me)
		if int(args.GetLeaderCommit()) > len(rf.logs)-1 {
			rf.commitIndex = int32(len(rf.logs) - 1)
		} else {
			rf.commitIndex = args.GetLeaderCommit()
		}
		rf.apply()
	}

	debug.Dlog("[Node %v]'s lastapplied log is %+v ,commited Index is %v", rf.me, rf.logs[rf.commitIndex], rf.commitIndex)
	reply.Success = true
	return reply, nil
}

func (rf *KVnode) RequestVote(ctx context.Context, args *RequestVoteArgs) (*RequestVoteReply, error) {
	reply := RequestVoteReply{}
	//debug.Dlog("[Node %v] receive a RequestVote from %v with %v.term is %v, %v's term is %v\n\t\t\tand rf.lastApply is %v, args.lastlogIndex is %v", rf.me, args.CANDIDATEID, args.CANDIDATEID, args.TERM, rf.me, rf.currentTerm, rf.lastApplied, args.LASTLOGINDEX)
	reply.Term = rf.currentTerm
	// Default not to vote
	reply.VoteGranted = false
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if rf.currentTerm > args.GetTerm() || (rf.currentTerm == args.GetTerm() && rf.votedFor != -1 && rf.votedFor != args.GetCandidateId()) || !rf.isLogUptoDate(int(args.LastLogIndex), int(args.LastLogIndex)) {
		// has voted to another Candidate
		debug.Dlog("[Node %v] DO NOT vote to %v, rf.votedFor is %v", rf.me, args.CandidateId, rf.votedFor)
		return &reply, nil
	}
	if rf.currentTerm <= args.Term {
		if rf.currentTerm == args.Term {
			if !rf.isLogUptoDate(int(args.LastLogIndex), int(args.LastLogTerm)) {
				return &reply, nil
			}
		}
		debug.Dlog("[Node %v] vote to %v", rf.me, args.CandidateId)
		reply.VoteGranted = true
		rf.votedFor = args.CandidateId
		rf.currentTerm = args.Term
		if rf.state == Candidate || rf.state == Leader {
			// go back to Follower
			reply.VoteGranted = true
			rf.state = Follower
		}
	}
	return &reply, nil
}

func (rf *KVnode) Operate(ctx context.Context, args *Operation) (*Opreturn, error) {
	reply := Opreturn{}

	return &reply, nil
}

// communicate with other servers
// have question here
func (rf *KVnode) SendAppendEntries(server int, args *AppendEntriesArgs) (*AppendEntriesReply, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	reply, succ := (*rf.stubs[server]).AppendEntries(ctx, args)
	if succ != nil {
		fmt.Printf("rpc call AppendEntries failed: %v\n", succ)
	}
	return reply, succ == nil
}
func (rf *KVnode) SendRequestVote(server int, args *RequestVoteArgs) (*RequestVoteReply, bool) {
	debug.Dlog("[Node %v] is sending RequestVote to %v:%v", rf.me, server, rf.conns[server].Target())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	reply, succ := (*rf.stubs[server]).RequestVote(ctx, args)
	if succ != nil {
		fmt.Printf("rpc call RequestVote failed: %v\n", succ)
		// 	arr := strings.Split(rf.addrs[server], ":")
		// 	cmd := exec.Command("telnet", arr[0], arr[1])
		// 	output, err := cmd.Output()
		// 	if err != nil {
		// 		fmt.Println("ping ", rf.addrs[server], " err:", err)
		// 	} else {
		// 		fmt.Println(string(output))
		// 	}
	}
	return reply, succ == nil
}

// used to create a new server for Register
func NewKVnode(me int, peers []int, addrs []string, applych chan ApplyMsg) *KVnode {
	debug.Dlog("[Node %v] enter newKVServer", me)
	iniEntry := &LogEntry{
		Term:  0,
		Index: 0,
	}
	logs := make([]*LogEntry, 0)
	logs = append(logs, iniEntry)
	stubs := make([]*RaftKVClient, len(peers))
	// connect to other servers
	conns := make([]*grpc.ClientConn, len(peers))

	rf := &KVnode{
		me:                int32(me),
		peers:             peers,
		addrs:             addrs,
		state:             Follower,
		logs:              logs,
		stubs:             stubs,
		conns:             conns,
		currentTerm:       0,
		votedFor:          -1,
		commitIndex:       0,
		lastApplied:       0,
		nextIndex:         make([]int32, len(peers)),
		matchIndex:        make([]int32, len(peers)),
		election_timeout:  time.NewTimer(RandElectionTimeout()),
		heartbeat_timeout: time.NewTimer(20 * time.Millisecond),
		applych:           applych,
		killed:            false,
	}

	//go rf.ticker()
	debug.Dlog("[Node %v] finished newKVnode", me)
	return rf
}

// connect with other
func (rf *KVnode) Connect() {
	//debug.Dlog("[Node %v] enter connect", rf.me)
	//stubs := make([]*raftKVClient, len(peers))
	for peer := range rf.peers {
		if peer == int(rf.me) {
			continue
		}
		debug.Dlog("[Node %v] is connecting with %v in %v", rf.me, peer, rf.addrs[peer])
		alive := keepalive.ClientParameters{
			Time:    1 * time.Minute,
			Timeout: 10 * time.Second,
		}
		conn, err := grpc.Dial(rf.addrs[peer], grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithKeepaliveParams(alive))
		rf.conns[peer] = conn
		if err != nil {
			panic(err)
		}
		stub := NewRaftKVClient(conn)
		rf.stubs[peer] = &stub
	}
	time.Sleep(10 * time.Millisecond)
	go rf.ticker()
	debug.Dlog("[Node %v] connect with others", rf.me)
}

// The ticker go routine starts a new election if this peer hasn't received
// heartsbeats recently, or send a heartbeat periodically if it is a Leader
func (rf *KVnode) ticker() {
	for !rf.killed {
		//debug.Dlog("[Node %v] enter ticker", rf.me)
		rf.mu.Lock()
		currentstate := rf.state
		rf.mu.Unlock()
		select {
		case <-rf.heartbeat_timeout.C:
			rf.mu.Lock()
			if currentstate == Leader {
				//debug.Dlog("[Node %v] send heart beart", rf.me)
				// send heartbeat to all followers
				rf.SendheartbeatToAll()
				rf.heartbeat_timeout.Reset(20 * time.Millisecond)
			}
			rf.mu.Unlock()
		case <-rf.election_timeout.C:
			rf.mu.Lock()
			rf.currentTerm += 1
			rf.state = Candidate
			// start an election event
			rf.StartElection()
			rf.election_timeout.Reset(RandElectionTimeout())
			rf.mu.Unlock()
		}
	}
}

func (rf *KVnode) StartElection() {
	debug.Dlog("[Node %v] start an election event", rf.me)
	rf.votedFor = rf.me
	logLen := len(rf.logs)
	agreeNum := 1 // itself
	// maybe there is no entry in log
	lastLogIndex := logLen - 1
	lastLogTerm := 0
	if logLen > 1 {
		//fmt.Printf("loglen is %v\n", logLen)
		lastLogTerm = int(rf.logs[logLen-1].Term)
	}
	args := &RequestVoteArgs{
		Term:         rf.currentTerm,
		CandidateId:  rf.me,
		LastLogTerm:  int32(lastLogTerm),
		LastLogIndex: int32(lastLogIndex),
	}
	agreeNumLock := sync.Mutex{}
	for server := range rf.peers {
		go func(peer int) {
			if peer != int(rf.me) {
				reply, succ := rf.SendRequestVote(peer, args)
				// receive the reply
				if succ {
					rf.mu.Lock()
					defer rf.mu.Unlock()
					if rf.currentTerm == args.GetTerm() && rf.state == Candidate {

						//debug.Dlog("[Node %v] Receive reply from %v", rf.me, peer)
						if reply.GetVoteGranted() {
							// get a vote
							debug.Dlog("[Node %v] get a vote from %v", rf.me, peer)
							agreeNumLock.Lock()
							agreeNum++

							// win this election and then send heartbeat, interrupt sending election message
							if agreeNum >= (len(rf.peers)+1)/2 {
								debug.Dlog("[Node %v] become a new leader", rf.me)
								rf.TobeLeader()
							}
							agreeNumLock.Unlock()
							return
						} else if reply.GetTerm() > rf.currentTerm {
							// find another Candidate/leader
							//debug.Dlog("[Node %v] find a larger term from %v", rf.me, peer)
							//rf.mu.Lock()
							//defer rf.mu.Unlock()
							rf.currentTerm = reply.GetTerm()
							// go back to Follower and interrupt sending message
							rf.state = Follower
							rf.election_timeout.Reset(RandElectionTimeout())
							rf.votedFor = -1
						}
					}
				}
			}
		}(server)
	}
}

func (rf *KVnode) TobeLeader() {
	rf.election_timeout.Reset(RandElectionTimeout())
	if len(rf.logs) > 0 {
		debug.Dlog("[Node %v] to be a leader with logs len: %v , term: %v and commitIndex: %v", rf.me, len(rf.logs), rf.currentTerm, rf.commitIndex)
	}
	rf.state = Leader
	lastIndex := len(rf.logs)
	for peer := range rf.nextIndex {
		rf.nextIndex[peer] = int32(lastIndex)
		rf.matchIndex[peer] = 0
	}

	rf.SendheartbeatToAll()
	rf.heartbeat_timeout.Reset(20 * time.Millisecond)
}

func (rf *KVnode) SendheartbeatToAll() {
	debug.Dlog("[Node %v] send heartbeat", rf.me)
	if rf.state == Leader {
		for server := range rf.peers {
			if server != int(rf.me) {
				prevIndex := rf.nextIndex[server] - 1
				debug.Dlog("[Node %v] send heartbeat to %v with prevIndex %v", rf.me, server, prevIndex)
				if prevIndex < 0 {
					debug.Dlog("[Node %v] don't send heartbeat to %v", rf.me, server)
					continue
				}
				prevTerm := rf.logs[prevIndex].Term
				entries := make([]*LogEntry, len(rf.logs)-int(prevIndex)-1)
				copy(entries, rf.logs[prevIndex+1:])
				args := &AppendEntriesArgs{
					Term:         rf.currentTerm,
					LeaderId:     rf.me,
					Entries:      entries,
					LeaderCommit: rf.commitIndex,
					PrevlogIndex: prevIndex,
					PrevLogTerm:  prevTerm,
				}

				go func(peer int) {
					reply, succ := rf.SendAppendEntries(peer, args)
					if succ {
						if reply.Success {
							rf.matchIndex[peer] = args.PrevlogIndex + int32(len(args.Entries))
							rf.nextIndex[peer] = rf.matchIndex[peer] + 1
							debug.Dlog("[Node %v] update %v nextIndex and matchIndex to %v %v", rf.me, peer, rf.nextIndex[peer], rf.matchIndex[peer])
							// update the commit index

							toCommit := make([]int, len(rf.logs))
							for index := range rf.peers {
								com := rf.matchIndex[index]
								toCommit[com]++
							}
							peerLen := len(rf.peers)
							// find the largest index which can be committed (at least larger than old commitIndex)
							sum := 0
							for i := len(toCommit) - 1; i > int(rf.commitIndex); i-- {
								sum += toCommit[i]
								if sum >= (1+peerLen)/2 {
									rf.commitIndex = int32(i)
									rf.apply()
									debug.Dlog("[Node %v] commitIndex is %v", rf.me, rf.commitIndex)
									break
								}
							}

						} else {
							// there is a client with larger term
							if reply.GetTerm() > rf.currentTerm {
								rf.state = Follower
								rf.election_timeout.Reset(RandElectionTimeout())
							} else {
								// there is no matching entry
								rf.nextIndex[peer] -= 1
								debug.Dlog("[Node %v] update %v nextIndex to %v", rf.me, peer, rf.nextIndex[peer])
							}
						}
					} else {
						// Loss of connection
						debug.Dlog("[Node %v] lost connection with %v", rf.me, peer)
					}
				}(server)
			}
		}
	}
}

func (rf *KVnode) SendNewCommandToAll() {
	debug.Dlog("[Node %v] is sending new command to others", rf.me)
	commitNum := 1
	commitNumLock := sync.Mutex{}
	oldCommit := rf.commitIndex
	for server := range rf.peers {
		if server == int(rf.me) {
			continue
		}
		go func(peer int) {
			prevterm := int32(0)
			if rf.nextIndex[peer]-1 > 0 {
				prevterm = rf.logs[rf.nextIndex[peer]-1].Term
			}
			entry := make([]*LogEntry, len(rf.logs[rf.nextIndex[peer]:]))
			args := &AppendEntriesArgs{
				Term:         rf.currentTerm,
				LeaderId:     rf.me,
				LeaderCommit: rf.commitIndex,
				PrevlogIndex: rf.nextIndex[peer] - 1,
				PrevLogTerm:  prevterm,
				Entries:      entry,
			}
			copy(args.Entries, rf.logs[rf.nextIndex[peer]:])

			debug.Dlog("[Node %v] the ENTRY with prevlogIndex %v for Node %v", rf.me, args.PrevlogIndex, peer)

			reply, succ := rf.SendAppendEntries(peer, args)
			if succ {

				if reply.GetTerm() > rf.currentTerm {
					rf.mu.Lock()
					rf.currentTerm = reply.GetTerm()
					rf.state = Follower
					rf.election_timeout.Reset(RandElectionTimeout())
					rf.votedFor = -1
					rf.mu.Unlock()
					return
				}
				if reply.Success {
					rf.mu.Lock()
					rf.matchIndex[peer] = args.PrevlogIndex + int32(len(args.Entries))
					rf.nextIndex[peer] = rf.matchIndex[peer] + 1
					debug.Dlog("[Node %v] update %v nextIndex and matchIndex to %v %v,and rf.commitIndex is %v，old is %v", rf.me, peer, rf.nextIndex[peer], rf.matchIndex[peer], rf.commitIndex, oldCommit)
					commitNumLock.Lock()
					commitNum++
					if rf.commitIndex == oldCommit && commitNum >= (len(rf.peers)+1)/2 {
						rf.commitIndex++
						debug.Dlog("[Node %v] commit a new command with commitId %v: %+v", rf.me, rf.commitIndex, rf.logs[rf.commitIndex])
						rf.apply()
						rf.SendheartbeatToAll()
						rf.heartbeat_timeout.Reset(20 * time.Millisecond)
					}
					commitNumLock.Unlock()
					rf.mu.Unlock()
				} else {
					rf.mu.Lock()
					rf.nextIndex[peer]--
					debug.Dlog("[Node %v] update %v nextIndex to %v", rf.me, peer, rf.nextIndex[peer])
					rf.mu.Unlock()
				}
			} else {
				debug.Dlog("[Node %v] lost the connection with %v", rf.me, peer)
			}
		}(server)
	}
}

// generate a rand election time
func RandElectionTimeout() time.Duration {
	source := rand.NewSource(time.Now().Local().UnixMicro())
	ran := rand.New(source)
	return time.Duration(400+ran.Int()%200) * time.Millisecond
}

func (rf *KVnode) isLogUptoDate(lastLogIndex int, lastLogTerm int) bool {
	if rf.lastApplied == 0 {
		return true
	} else {
		debug.Dlog("[Node %v] checking isn't logUptoDate: lastLogIndex is %v, lastLogTerm is %v", rf.me, lastLogIndex, lastLogTerm)
		//if rf.logs[rf.lastApplied].Term > lastLogTerm || rf.lastApplied > lastLogIndex {
		//	return false
		//}
		if rf.logs[len(rf.logs)-1].Term > int32(lastLogTerm) {
			return false
		}
		if rf.logs[len(rf.logs)-1].Term == int32(lastLogTerm) {
			if len(rf.logs)-1 > lastLogIndex {
				return false
			}
		}
	}
	return true
}

func (rf *KVnode) IsLeader() bool {
	return rf.state == Leader
}

func (rf *KVnode) apply() {
	for rf.commitIndex > rf.lastApplied && rf.lastApplied+1 < int32(len(rf.logs)) {
		rf.lastApplied++
		op := rf.logs[rf.lastApplied].GetOp()
		key := rf.logs[rf.lastApplied].GetKey()
		value := rf.logs[rf.lastApplied].GetValue()

		applymsg := ApplyMsg{
			Op:    op,
			Key:   key,
			Value: value,
			Index: int(rf.lastApplied),
		}
		//debug.Dlog("[Node %v] is applying %v: %+v\n\t\t\t\tSending msg to applych", rf.me, rf.lastApplied, rf.logs[rf.lastApplied])
		rf.applych <- applymsg
		debug.Dlog("[Node %v] send msg to ch successfully with index:%v", rf.me, applymsg.Index)
	}
}

func (rf *KVnode) Exec(op string, key string, value int32) (int, int, bool) {
	index := -1
	term := -1
	isLeader := rf.IsLeader()
	if isLeader {
		rf.mu.Lock()
		term = int(rf.currentTerm)
		index = len(rf.logs)
		rf.logs = append(rf.logs, &LogEntry{
			Term:  int32(term),
			Index: int32(index),
			Op:    op,
			Key:   key,
			Value: int32(value),
		})
		rf.SendNewCommandToAll()
		rf.mu.Unlock()
	}
	return index, term, isLeader
}

func (rf *KVnode) Close() error {
	for {
		rf.mu.Lock()
		if int(rf.lastApplied) == int(rf.commitIndex) {
			debug.Dlog("[Node %v] finished applying all committed logs", rf.me)
			rf.election_timeout.Stop()
			rf.heartbeat_timeout.Stop()
			rf.mu.Unlock()
			break
		}
		rf.mu.Unlock()
		time.Sleep(5 * time.Millisecond)
	}
	for i := range rf.conns {
		if int32(i) == rf.me {
			continue
		}
		err := rf.conns[i].Close()
		if err != nil {
			panic(err)
		}
	}
	return nil
}

package server

import (
	// "project/db"
	"sync"
	"time"

	"github.com/darshan-na/MetaStore/db"
)

type LogEntry struct {
	Term    int
	Command interface{}
}

type State int

const (
	Follower State = iota
	Candidate
	Leader
)

type Raft struct {
	mu             sync.RWMutex
	peers          []string
	me             int
	currentTerm    int
	votedFor       int
	log            []LogEntry
	commitIndex    int
	lastApplied    int
	nextIndex      []int
	matchIndex     []int
	state          State
	voteCount      int
	electionTimer  *time.Timer
	heartbeatTimer *time.Timer
	db             *db.DB
}

func NewRaft() *Raft {
	// Implementation
	return &Raft{peers: make([]string, 0, 5)}
}

func (rf *Raft) GetPeers() []string {
	rf.mu.RLock()
	defer rf.mu.RUnlock()
	peersCopy := make([]string, len(rf.peers))
	copy(peersCopy, rf.peers)
	return peersCopy
}

func (rf *Raft) SetPeer(hostAddr string) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	rf.peers = append(rf.peers, hostAddr)
}

func (rf *Raft) Run() {
	// Implementation
}

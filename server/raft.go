package server

import (
	"project/db"
	"sync"
	"time"
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
	mu             sync.Mutex
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

func NewRaft(peers []string, me int) *Raft {
	// Implementation
	return &Raft{}
}

func (rf *Raft) Run() {
	// Implementation
}

package base

import (
	"fmt"
	"net"
	"time"
)

const UrlDelimiter = "/"
const AddPeer = "addPeer"
const SetPeer = "setPeer"
const GetPeers = "getPeers"
const HttpPrefix = "http://"
const HttpTimeout = 100 * time.Second
const TCP = "tcp"
const NewPeerAddr = "NewPeerAddr"
const MyPeers = "Peers"

type ErrorMap map[string]error
type StatusMap map[string]int

var (
	Dialer *net.Dialer = &net.Dialer{Timeout: HttpTimeout}
)

var ErrorDoesNotExistString = "does not exist"
var ErrorResourceDoesNotExist = fmt.Errorf("specified resource %v", ErrorDoesNotExistString)

const (
	GetMethod    = "GET"
	PostMethod   = "POST"
	DeleteMethod = "DELETE"
)

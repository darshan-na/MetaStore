package network

import (
	_ "expvar"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/darshan-na/MetaStore/server"
)

type ContentType int

const (
	ContentTypeJson ContentType = iota
	ContentTypeText ContentType = iota
)

func (ct ContentType) String() string {
	switch ct {
	case ContentTypeJson:
		return "application/json"
	case ContentTypeText:
		return "text/plain"
	default:
		panic("Invalid Type")
	}
}

type Request interface {
	// Get http request from request packet.
	GetRequest() *http.Request

	// Send a response back to the client.
	SendResponse(response *Response) error

	// Send error back to the client.
	SendError(error) error

	// get http server for this request
	GetServer() *httpServer
}

type httpRequest struct {
	srv    *httpServer      //server that receives the request
	req    *http.Request    //the actual request
	waitch chan interface{} //for communicating the response to the client
}

// GetHttpRequest is part of Request interface.
func (r *httpRequest) GetRequest() *http.Request {
	return r.req
}

func (r *httpRequest) GetServer() *httpServer {
	return r.srv
}

// Send is part of Request interface.
func (r *httpRequest) SendResponse(response *Response) error {
	r.waitch <- response
	close(r.waitch)
	return nil
}

// SendError is part of Request interface.
func (r *httpRequest) SendError(err error) error {
	r.waitch <- err
	close(r.waitch)
	return nil
}

type Response struct {
	StatusCode  int
	Body        []byte
	ContentType ContentType
}

type Server interface {
	Start() chan error
	Stop()
}
type httpServer struct {
	lock     sync.RWMutex
	srv      *http.Server   // http server
	reqch    chan<- Request // request channel back to application
	raftNode *server.Raft
}

func (s *httpServer) Start() chan error {
	errCh := make(chan error, 1)

	// Server routine
	go func() {
		defer s.shutdown()
		// ListenAndServe blocks and returns a non-nil error if something wrong happens
		// ListenAndServe will cause golang library to call ServeHttp()
		err := s.srv.ListenAndServe()
		errCh <- err
	}()
	return errCh
}
func (s *httpServer) shutdown() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.srv != nil {
		s.srv.Close()
		close(s.reqch)
		s.srv = nil
	}
}

// Stop is part of Server interface. Once stopped, Start() cannot be called again
func (s *httpServer) Stop() {
	s.shutdown()
}

func (s *httpServer) GetRaftNode() *server.Raft {
	return s.raftNode
}

func (s *httpServer) systemHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	// Fault-tolerance. No need to crash the server in case of panic.
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("adminport.request.recovered `%v`\n", r)
		} else if err != nil {
			fmt.Printf("%v\n", err)
		}
	}()

	actionStr := fmt.Sprintf("Method=%s, URL=%s", r.Method, r.URL)
	ExecWithTimeout3(
		actionStr,

		// Call this closure first as action
		func(waitch chan interface{}) error {
			s.reqch <- &httpRequest{srv: s, req: r, waitch: waitch}
			return nil
		},

		// Call this closure if response is received within timeout
		func(val interface{}) error {
			switch v := (val).(type) {
			case error:
				http.Error(w, v.Error(), http.StatusInternalServerError)
				err := fmt.Errorf("%v, %v", "Internal error in adminport", v)
				fmt.Printf("%v", err)
			case *Response:
				w.Header().Set("Content-Type", v.ContentType.String())
				w.WriteHeader(v.StatusCode)
				w.Write(v.Body)
				return nil
			}
			return nil
		},

		30*time.Second)

}

type Handler struct {
	server *httpServer
}

// Called by golang library
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.server.systemHandler(w, r)
}

func (h *Handler) SetServer(s Server) error {
	server, ok := s.(*httpServer)
	if !ok {
		return fmt.Errorf("Error")
	}
	h.server = server
	return nil
}

func (h *Handler) GetServer() Server {
	return h.server
}

// Start is part of Server interface.

// ExecWithTimeout3 is a variant which executes action and if response is received
// within timeout rspHandler is called. actionStr is the contextual info which is
// logged when timeout occurs. The action should be spawned in a go-routine.
func ExecWithTimeout3(actionStr string,
	action func(finch chan interface{}) error,
	rspHandler func(val interface{}) error,
	timeoutDur time.Duration) (err error) {

	waitch := make(chan interface{}, 1)
	// We do not close the waitch channel
	// This is because some other goroutine (in action) might be writing to it
	// while this function potentially closing it, which may cause panic

	errch := make(chan error, 1)

	go func(ch chan error) {
		err = action(waitch)
		if err != nil {
			errch <- err
		}
	}(errch)

	timeoutticker := time.NewTicker(timeoutDur)
	defer timeoutticker.Stop()

	select {
	case <-timeoutticker.C:
		fmt.Printf("Executing Action timed out. action: %s", actionStr)
		fmt.Printf("****************************")
		fmt.Printf("ErrorExecutionTimedOut")
	case err = <-errch:
		return err
	case val := <-waitch:
		rspHandler(val)
	}

	return
}

func NewHttpServer(handler http.Handler, reqch chan<- Request) *httpServer {
	//0.0.0.0 indicates that the server can accept connections from any IP addresses and is listening on port 7070
	return &httpServer{srv: &http.Server{Addr: "0.0.0.0:7070", Handler: handler, MaxHeaderBytes: 1 << 20}, reqch: reqch, raftNode: server.NewRaft()}
}

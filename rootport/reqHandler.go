package rootport

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/darshan-na/MetaStore/base"
	"github.com/darshan-na/MetaStore/network"
	"github.com/darshan-na/MetaStore/utils"
)

func ReqSerializer(reqs <-chan network.Request, finch <-chan bool) {
	for {
		select {
		case <-finch:
			fmt.Printf("Darshan: Exiting the req serilizer. Probably the server stopped\n")
			return
		case req, ok := <-reqs:
			if !ok { //channel closed
				return
			}
			HandleRequest(req)
			// fmt.Printf("Darshan: Logic to handle the req should go here in future. req is on path %v\n", req.GetRequest().URL)
			// req.SendResponse(&Response{StatusCode: 200, Body: []byte("Success"), ContentType: ContentTypeText})
		}

	}
}

func validateKey(key string) error {
	if key != "hostAddr" {
		return fmt.Errorf("Invalid key %v", key)
	}
	return nil
}

func doGetPeers(reqIface network.Request) (*network.Response, error) {
	raftNode := reqIface.GetServer().GetRaftNode()
	peers := raftNode.GetPeers()
	return &network.Response{StatusCode: 200, Body: []byte(fmt.Sprintf("Peers: %v", peers)), ContentType: network.ContentTypeText}, nil
}

func doSetPeer(reqIface network.Request) (*network.Response, error) {
	request := reqIface.GetRequest()
	raftNode := reqIface.GetServer().GetRaftNode()
	reqBody := make(map[string]interface{})
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &reqBody)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Darshan Hello %v\n", reqBody)
	NewPeer, ok := reqBody[base.NewPeerAddr].(string)
	if ok {
		raftNode.SetPeer(NewPeer)
	}
	fmt.Printf("Darshan Type %T\n", reqBody[base.MyPeers])
	myPeers, ok1 := reqBody[base.MyPeers].([]interface{})
	if ok1 {
		for _, peer := range myPeers {
			raftNode.SetPeer(peer.(string))
		}
	}
	if !ok && !ok1 {
		return nil, fmt.Errorf("failed to set peer")
	} else {
		return &network.Response{StatusCode: 200, Body: []byte(fmt.Sprintf("Successfully set peer %v on %v", utils.GetLocalIP())), ContentType: network.ContentTypeText}, nil
	}
}

func doAddPeer(reqIface network.Request) (*network.Response, error) {
	request := reqIface.GetRequest()
	raftNode := reqIface.GetServer().GetRaftNode()
	if err := request.ParseForm(); err != nil {
		return nil, err
	}
	var childWaitGrp sync.WaitGroup
	var newPeerAddr string
	for key, val := range request.Form {
		err := validateKey(key)
		if err != nil {
			return nil, fmt.Errorf("Failed to add node to the cluster. err=%v", err)
		}
		newPeerAddr = val[0]
		break
	}
	errMap := make(base.ErrorMap)
	statusMap := make(base.StatusMap)
	// 1. For every current peer send the new node
	// 2. Send currentPeers+me to the new node
	// 3. Update the peers on the current node
	peers := raftNode.GetPeers()
	for _, hostAddr := range peers {

		body, err := json.Marshal(map[string]string{
			base.NewPeerAddr: newPeerAddr,
		})
		if err != nil {
			fmt.Printf("Failed to send AddPeerReq to Host %v\n", hostAddr)
		}
		childWaitGrp.Add(1)
		go utils.SendRestReq(hostAddr, base.SetPeer, body, base.PostMethod, nil, errMap, statusMap, &childWaitGrp)
	}
	peersPlusMe := append(peers, utils.GetLocalIP())
	fmt.Printf("Darshan: IP %v\n", utils.GetLocalIP())
	body, err := json.Marshal(map[string]interface{}{
		base.MyPeers: peersPlusMe,
	})
	if err != nil {
		fmt.Printf("Failed to send AddPeerReq to Host %v\n", newPeerAddr)
	}
	childWaitGrp.Add(1)
	go utils.SendRestReq(newPeerAddr, base.SetPeer, body, base.PostMethod, nil, errMap, statusMap, &childWaitGrp)
	raftNode.SetPeer(newPeerAddr)
	childWaitGrp.Wait()
	if len(errMap) > 0 {
		return nil, fmt.Errorf("failed to addPeer %v err=%v", newPeerAddr, utils.FlattenErrorMap(errMap))
	}
	return &network.Response{StatusCode: 200, Body: []byte(fmt.Sprintf("Successfully added peer %v to the cluster", newPeerAddr)), ContentType: network.ContentTypeText}, nil
}

func HandleRequest(req network.Request) {
	path := utils.GetMessageFromUrl(req)
	fmt.Printf("Path is %v\n", path)
	switch path {
	case base.AddPeer + base.UrlDelimiter + base.PostMethod:
		resp, err := doAddPeer(req)
		if err != nil {
			req.SendError(err)
		} else {
			req.SendResponse(resp)
		}
	case base.SetPeer + base.UrlDelimiter + base.PostMethod:
		resp, err := doSetPeer(req)
		if err != nil {
			req.SendError(err)
		} else {
			req.SendResponse(resp)
		}
	case base.GetPeers + base.UrlDelimiter + base.GetMethod:
		resp, err := doGetPeers(req)
		if err != nil {
			req.SendError(err)
		} else {
			req.SendResponse(resp)
		}
	}
}

package main

import (
	"fmt"

	"github.com/darshan-na/MetaStore/network"
	"github.com/darshan-na/MetaStore/rootport"
)

var donech = make(chan bool, 1)

func main() {
	fmt.Printf("Hello 1\n")
	handler := network.Handler{}
	reqch := make(chan network.Request, 100)
	server := network.NewHttpServer(&handler, reqch)
	handler.SetServer(server)
	server.Start()
	go rootport.ReqSerializer(reqch, make(<-chan bool))
	//use the below code snippet to test the http server
	// go func() {
	// 	fmt.Printf("darshan helo\n")
	// 	for {
	// 		select {
	// 		case <-finch:
	// 			server.Stop()
	// 			return
	// 		case req, ok := <-reqch:
	// 			if !ok {
	// 				return
	// 			} else {
	// 				fmt.Printf("Darshan: Received Request %v\n", req.GetRequest().URL)
	// 				req.SendResponse(&network.Response{StatusCode: 200, Body: []byte("Success"), ContentType: network.ContentTypeText})
	// 			}
	// 		}
	// 	}
	// }()
	<-donech
}

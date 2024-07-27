package network

import (
	"fmt"
)

func ReqSerializer(reqs <-chan Request, finch <-chan bool) {
	for {
		select {
		case <-finch:
			fmt.Printf("Darshan: Exiting the req serilizer. Probably the server stopped\n")
			return
		case req, ok := <-reqs:
			if !ok { //channel closed
				return
			}
			fmt.Printf("Darshan: Logic to handle the req should go here in future. req is on path %v\n", req.GetRequest().URL)
			req.SendResponse(&Response{StatusCode: 200, Body: []byte("Success"), ContentType: ContentTypeText})
		}

	}
}

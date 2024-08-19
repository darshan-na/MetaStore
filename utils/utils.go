package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/darshan-na/MetaStore/base"
	"github.com/darshan-na/MetaStore/network"
)

func GetMessageFromUrl(req network.Request) string {
	path := req.GetRequest().URL.Path[len(base.UrlDelimiter):]
	fmt.Printf("specified path in the request is %v\n", path)
	if strings.HasSuffix(path, base.UrlDelimiter) {
		path = path[:len(path)-1]
	}
	if len(path) != 0 {
		path += base.UrlDelimiter + strings.ToUpper(req.GetRequest().Method)
	}
	return path
}

func GetLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		fmt.Printf("Failed to fetch local Ip Addr. err=%v", err)
	}
	defer conn.Close()

	localAddress := conn.LocalAddr().(*net.UDPAddr)
	return localAddress.IP.String() + ":7070"
}

func PrepareUrlForRequest(hostAddr, path string) string {
	return base.HttpPrefix + hostAddr + base.UrlDelimiter + path
}

func PrepareForRestCall(hostAddr, method, path string, body []byte) (*http.Client, *http.Request, error) {

	client := &http.Client{Timeout: base.HttpTimeout}
	url := PrepareUrlForRequest(hostAddr, path)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, nil, err
	}
	return client, req, nil

}

func SendRestReq(hostAddr string, path string, body []byte, method string, result interface{}, errMap base.ErrorMap, statusMap base.StatusMap, wg *sync.WaitGroup) {
	defer wg.Done()
	var client *http.Client
	var req *http.Request
	var resp *http.Response
	var err error
	client, req, err = PrepareForRestCall(hostAddr, method, path, body)
	if err != nil {
		errMap[hostAddr] = err
		statusMap[hostAddr] = 0
		return
	}
	resp, err = client.Do(req)
	if err != nil {
		errMap[hostAddr] = err
		statusMap[hostAddr] = 0
		return
	}
	err = ParseResponse(resp, result)
	if err != nil {
		errMap[hostAddr] = err
		statusMap[hostAddr] = resp.StatusCode
	}
}

func ParseResponse(resp *http.Response, result interface{}) (err error) {
	if resp != nil && resp.Body != nil {
		// response.Body is streamed on demand. Hence it is the callers responsibilty to close the stream once done else we'll be leaking resources
		defer resp.Body.Close()
		var bod []byte
		if resp.ContentLength == 0 {
			// If res.Body is empty, json.Unmarshal on an empty Body will return the error "unexpected end of JSON input"
			err = base.ErrorResourceDoesNotExist
			return
		}
		bod, err = io.ReadAll(io.LimitReader(resp.Body, resp.ContentLength))
		if err != nil {
			err = fmt.Errorf("failed to read response body, err=%v\n res=%v", err, resp)
			return
		}
		if result != nil {
			err = json.Unmarshal(bod, result)
		}
	}
	return
}

func FlattenErrorMap(errMap base.ErrorMap) string {
	var buffer bytes.Buffer
	var first bool = true
	for k, v := range errMap {
		if !first {
			buffer.WriteByte('\n')
		} else {
			first = false
		}
		buffer.WriteString(k)
		buffer.WriteString(" : ")
		if v == nil {
			buffer.WriteString("nil")
		} else {
			buffer.WriteString(v.Error())
		}
	}
	return buffer.String()
}

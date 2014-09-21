package haproxy

import (
	"io"
	"net"
)

type HAProxyRequest struct {
	Command string
}

type HAProxyReply struct {
	Message string
}

type HAProxyClient struct {
	socketPath string
}

func NewClient(pathToSocket string) *HAProxyClient {
	return &HAProxyClient{socketPath: pathToSocket}
}

func NewRequest(command string) (*HAProxyRequest, error) {
	request := new(HAProxyRequest)
	// TODO : check these are valid requests
	request.Command = command
	return request, nil
}

func (client *HAProxyClient) Rpc(command string) (*HAProxyReply, error) {
	request, _ := NewRequest(command)
	return doRpc(client, request)
}

func doRpc(client *HAProxyClient, request *HAProxyRequest) (*HAProxyReply, error) {

	buf := make([]byte, 512)

	socket := client.socketPath
	sockettype := "unix"
	conn, err := net.Dial(sockettype, socket)
	if err != nil {
		return new(HAProxyReply), err
	}
	defer conn.Close()

	_, err = conn.Write([]byte(request.Command))
	if err != nil {
		return new(HAProxyReply), err
	}

	n, err := conn.Read(buf)
	resp := make([]byte, n)

	switch err {
	case io.EOF:
		resp = append(resp, buf[:n]...)
	case nil:
		resp = append(resp, buf[:n]...)
	default:
		return new(HAProxyReply), err
	}

	toreturn := new(HAProxyReply)
	toreturn.Message = string(resp[:])
	return toreturn, nil
}

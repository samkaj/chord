package chord

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

type Empty struct{}

type FindSuccessorArgs struct {
	Key string
}

type FindSuccessorReply struct {
	Successor string
}

type GetPredecessorReply struct {
	Predecessor string
}

type NotifyArgs struct {
	Key string
}

type NotifyReply struct {
	Success bool
}

type ClosestPrecedingNodeArgs struct {
	Key string
}

type ClosestPrecedingNodeReply struct {
	Node string
}

type GetSuccessorlistArgs struct{}
type GetSuccessorlistReply struct {
	Successors []string
}

func (node *Node) ServeAndListen() {
	rpc.Register(node)
	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", node.Address)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	fmt.Printf("Listening on %s\n", node.Address)
	err = http.Serve(listener, nil)
}

func call(method string, address string, args any, reply any) error {
	conn, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		return fmt.Errorf("Failed to dial: %v", err)
	}

	err = conn.Call(method, args, reply)
	if err != nil {
		return fmt.Errorf("Failed to call: %v", err)
	}
	conn.Close()
	return nil
}

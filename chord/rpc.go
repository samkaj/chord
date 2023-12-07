package chord

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

type Empty struct {}

type FindSuccessorArgs struct {
  CallingNode *Node
}

type FindSuccessorReply struct {
	Successor string
}

type GetPredecessorReply struct {
    Predecessor string
}

type NotifyArgs struct {
  CallingNode *Node
}

type NotifyReply struct {
  Success bool
}

type ClosestPrecedingNodeArgs struct {
  CallingNode *Node
}

type ClosestPrecedingNodeReply struct {
  Node string
}

type PingReply struct {
    Alive bool
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

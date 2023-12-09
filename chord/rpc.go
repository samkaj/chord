package chord

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

type Empty struct{}

type NodeRef struct {
	Address    string
	PublicKey  []byte
	TLSAddress string
}

type FindSuccessorArgs struct {
	Key string
}

type FindSuccessorReply struct {
	Successor NodeRef
}

type GetPredecessorReply struct {
	Predecessor NodeRef
}

type NotifyArgs struct {
	Key NodeRef
}

type NotifyReply struct {
	Success bool
}

type ClosestPrecedingNodeArgs struct {
	Key string
}

type ClosestPrecedingNodeReply struct {
	Node NodeRef
}

type GetSuccessorlistArgs struct{}
type GetSuccessorlistReply struct {
	Successors []NodeRef
}

type StoreFileArgs struct {
  Path string
  Data []byte
}

type StoreFileReply struct {
  Success bool
}

func (node *Node) ServeAndListen() {
	rpc.Register(node)
	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", node.Address)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("Listening on %s\n", node.Address)
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

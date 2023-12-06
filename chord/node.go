package chord

import (
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

const r int = 3
const maxRequests int = 32

type Node struct {
	ID                       *big.Int
	Address                  string
	Successor                []string
	Predecessor              *Node
	StabilizeInterval        int
	FixFingersInterval       int
	CheckPredecessorInterval int
}

func (n *Node) CreateRing() {
	n.Predecessor = nil
	n.Successor = []string{n.Address} // FIXME: should call find successor on the dest node
	n.ID = hash(n.Address)
	n.Start()
}

// Joins a ring and sets the destination to the successor of the node.
func (n *Node) JoinRing(dest string) {
	n.Predecessor = nil
	n.Successor = []string{dest} // FIXME: should call find successor on the dest node
	n.ID = hash(n.Address)
	n.Start()
}

// Calls the functions on their specified intervals.
func (n *Node) StartIntervals() {
	go callOnInverval(n.CheckPredecessorInterval, n.CheckPredecessor)
	go callOnInverval(n.StabilizeInterval, n.Stabilize)
	go callOnInverval(n.FixFingersInterval, n.FixFingers)
}

// Call the successor and get its successor list and predecessor.
func (n *Node) Stabilize() {
	log.Println("successors: ", len(n.Successor))
	// Don't call yourself
	if len(n.Successor) < 1 {
		n.Successor = []string{n.Address}
	}

	if n.Successor[0] == n.Address {
		return
	}

	alive := &PingReply{}
	err := call("Node.Ping", n.Successor[0], Empty{}, alive)
	if err != nil {
		log.Println(err)
		if len(n.Successor) > 0 {
			n.Successor = n.Successor[1:]
		}
	}

	reply := &NotifyReply{}
	if len(n.Successor) == 0 {
		n.Successor = []string{n.Address}
		return
	}
	err = call("Node.Notify", n.Successor[0], &NotifyArgs{Node: *n}, reply)
	successors := []string{n.Successor[0]}
	successors = append(successors, reply.Successors...)
	n.Successor = successors
}

func (n *Node) Notify(args *NotifyArgs, reply *NotifyReply) error {
	if equals(n.ID, args.Node.ID) {
		return nil
	}
	if n.Predecessor == nil /*|| (len(n.Successor) > 0 && between(n.Predecessor.ID, args.Node.ID, n.Successor[0].ID, false) )*/ {
		n.Predecessor = &args.Node
	}
	reply.Predecessor = *n.Predecessor
	reply.Successors = n.Successor
	return nil
}

func (n *Node) FixFingers() {}
func (n *Node) CheckPredecessor() {
	if n.Predecessor != nil {
		fmt.Println("check predecessor")
		reply := &PingReply{}
		err := call("Node.Ping", n.Predecessor.Address, Empty{}, reply)
		if !reply.Alive || err != nil {
			n.Predecessor = nil
		}
	}
}

func (n *Node) Ping(args *Empty, reply *PingReply) error {
	fmt.Println("ping")
	reply.Alive = true
	return nil
}

func callOnInverval(ms int, f func()) {
	for {
		f()
		time.Sleep(time.Duration(ms) * time.Millisecond)
	}
}

// Launch the RPC server and perform other necessary stuff
func (n *Node) Start() {
	n.StartIntervals()
	n.LaunchRPC()
}

// Launches an RPC server on the node's address
func (n *Node) LaunchRPC() {
	rpc.Register(n)
	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", n.Address)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
		return
	}

	fmt.Println("Starting RPC server on ", n.Address)
	go http.Serve(listener, nil)
}

// -------------------- RPC start --------------------

// Calls the specified function via RPC
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

type Empty struct{}

type PingReply struct {
	Alive bool
}

type NotifyReply struct {
	Predecessor Node
	Successors  []string
}

type NotifyArgs struct {
	Node Node
}

// -------------------- RPC end ----------------------

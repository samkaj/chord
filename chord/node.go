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
}

// Joins a ring and sets the destination to the successor of the node.
func (n *Node) JoinRing(dest string) {
}

// Calls the functions on their specified intervals.
func (n *Node) StartIntervals() {
	go callOnInverval(n.CheckPredecessorInterval, n.CheckPredecessor)
	go callOnInverval(n.StabilizeInterval, n.Stabilize)
	go callOnInverval(n.FixFingersInterval, n.FixFingers)
}

// Call the successor and get its successor list and predecessor.
func (n *Node) Stabilize() {
}

func (n *Node) Notify(args *NotifyArgs, reply *NotifyReply) error {
  return nil
}

func (n *Node) FixFingers() {}
func (n *Node) CheckPredecessor() {
}

func (n *Node) ClosestPreceedingNode(id string) string {
  return ""
}

func (n *Node) Find(args *FindArgs, reply *FindReply) error {
	return fmt.Errorf("could not find successor")
}

func (n *Node) FindSuccessor(id string) (bool, string) {
  return false, ""
}

func (n *Node) Ping(args *Empty, reply *PingReply) error {
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

type FindArgs struct {
	ID    string
	Start string
}

type FindReply struct {
	Found bool
	ID    string
}

// -------------------- RPC end ----------------------

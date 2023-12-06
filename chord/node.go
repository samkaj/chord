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

// Create a new chord ring.
func (n *Node) CreateRing() {
	n.Predecessor = nil
	n.Successor = make([]string, r)
	n.Successor[0] = n.Address
	n.Start()
}

// Join a chord ring containing node dest.
func (n *Node) JoinRing(dest string) {
	n.Predecessor = nil
	n.Successor = make([]string, r)
	n.Successor[0] = dest
	n.Start()
}

type NotifyReply struct {
	Predecessor Node
	Successors  []string
}

type NotifyArgs struct {
	Node Node
}

func (n *Node) Notify(args *NotifyArgs, reply *NotifyReply) error {
  log.Println("Notify")
	if n.Predecessor == nil || between(args.Node.ID, n.Predecessor.ID, n.ID, false) {
		n.Predecessor = &args.Node
	}
	return nil
}

// Verify the node's successor and notify it of this node's existence.
func (n *Node) Stabilize() {
  if n.Successor == nil {
    return
  }
	successor := n.Successor[0]
	var reply NotifyReply
	err := call("Node.Notify", successor, &NotifyArgs{*n}, &reply)
	if err != nil {
		fmt.Println("Failed to stabilize: ", err)
		return
	}
	n.Successor = reply.Successors
}

func (n *Node) FixFingers() {}

func (n *Node) CheckPredecessor() {
	if n.Predecessor == nil {
		return
	}
	var reply PingReply
	err := call("Node.Ping", n.Predecessor.Address, &Empty{}, &reply)
	if err != nil || !reply.Alive {
		n.Predecessor = nil
	}
}

func (n *Node) FindSuccessor(id string) string {
	if between(toBigInt(id), n.ID, hash(n.Successor[0]), true) {
		return n.Successor[0]
	}
	closest := n.ClosestPreceedingNode(id)
	if closest == n.Address {
		return n.Successor[0]
	}
	var reply FindReply
	err := call("Node.Find", closest, &FindArgs{id, n.Address}, &reply)
	if err != nil {
		fmt.Println("Failed to find successor: ", err)
		return ""
	}
	if reply.Found {
		return reply.ID
	}
	return n.Successor[0]
}

func (n *Node) ClosestPreceedingNode(id string) string {
	return n.Address
}

func (n *Node) Find(args *FindArgs, reply *FindReply) error {
  log.Println("Find")
	return fmt.Errorf("could not find successor")
}

func (n *Node) Ping(args *Empty, reply *PingReply) error {
  log.Println("Ping")
	reply.Alive = true
	return nil
}

// Calls the functions on their specified intervals.
func (n *Node) StartIntervals() {
	go callOnInverval(n.CheckPredecessorInterval, n.CheckPredecessor)
	go callOnInverval(n.StabilizeInterval, n.Stabilize)
	go callOnInverval(n.FixFingersInterval, n.FixFingers)
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

type FindArgs struct {
	ID    string
	Start string
}

type FindReply struct {
	Found bool
	ID    string
}

// -------------------- RPC end ----------------------

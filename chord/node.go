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
	Successor                []*Node
	Predecessor              *Node
	StabilizeInterval        int
	FixFingersInterval       int
	CheckPredecessorInterval int
}

// Create a new chord ring.
func (n *Node) CreateRing() {
	n.Predecessor = nil
	n.Successor = make([]*Node, r)
	n.Successor[0] = n
	n.Start()
}

// Join a chord ring containing node dest.
func (n *Node) JoinRing(dest string) {
	n.Predecessor = nil
	n.Successor = make([]*Node, r)
	reply := new(FindReply)
	err := call("Node.FindSuccessor", dest, &FindArgs{n.Address}, reply)
	if err != nil {
		log.Fatalf("Failed to join ring: %v", err)
	}
	n.Successor[0] = reply.FoundNode
	n.Start()
}

type FindArgs struct {
	AddressToFind string
}

type FindReply struct {
	Found     bool
	FoundNode *Node
}

// Ask the node to find the successor of the specified ID.
func (n *Node) FindSuccessor(args *FindArgs, reply *FindReply) error {
	log.Println("Find")
	if len(n.Successor) > 0 && between(hash(n.Address), toBigInt(args.AddressToFind), n.Successor[0].ID, true) {
		reply.Found = true
		reply.FoundNode = n.Successor[0]
	} else {
		reply.Found = false
		reply.FoundNode = n.ClosestPreceedingNode(args.AddressToFind)
		nArgs := new(FindArgs)
		nArgs.AddressToFind = args.AddressToFind
		nReply := new(FindReply)
		err := call("Node.FindSuccessor", reply.FoundNode.Address, nArgs, nReply)
		if err != nil {
			return err
		}
	}
	return nil
}

type NotifyReply struct {
	Predecessor Node
	Successors  []string
}

type NotifyArgs struct {
	Node Node
}

func (n *Node) Notify(args *NotifyArgs, reply *NotifyReply) error {
	if n.Predecessor == nil || between(hash(n.Predecessor.Address), hash(args.Node.Address), hash(n.Address), false) {
		n.Predecessor = &args.Node
	}
	return nil
}

// Verify the node's successor and notify it of this node's existence.
func (n *Node) Stabilize() {
	x := n.Successor[0].Predecessor
	if x != nil && between(hash(n.Address), hash(x.Address), hash(n.Successor[0].Address), false) {
		n.Successor[0] = x
	}

	nArgs := new(NotifyArgs)
	nArgs.Node = *n
	nReply := new(NotifyReply)
	// don't call notify on self
	if n.Successor[0].Address == n.Address {
		return
	}
	err := call("Node.Notify", n.Successor[0].Address, nArgs, nReply)
	if err != nil {
		log.Fatalf("Failed to stabilize: %v", err)
	}
}

func (n *Node) FixFingers() {}

func (n *Node) CheckPredecessor() {
	pingArgs := new(Empty)
	pingReply := new(PingReply)
	if n.Predecessor == nil {
		return
	}
	err := call("Node.Ping", n.Predecessor.Address, pingArgs, pingReply)
	if err != nil {
		n.Predecessor = nil
	}
}

func (n *Node) ClosestPreceedingNode(id string) *Node {
	return n.Successor[0]
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

// -------------------- RPC end ----------------------

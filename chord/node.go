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
	n.ID = hash(n.Address)
	n.Successor = []string{}
	n.Start()
}

// Joins a ring and sets the destination to the successor of the node.
func (n *Node) JoinRing(dest string) {
	n.Predecessor = nil

	n.ID = hash(n.Address)
	//reply := &FindReply{}
	//call("Node.Find", dest, &FindArgs{ID: n.ID.String(), Start: dest}, reply)
	n.Successor = []string{dest}
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
	log.Println("predecessor: ", n.Predecessor)
	if len(n.Successor) == 0 {
		n.Successor = []string{n.Address}
	}

	if n.Successor[0] != n.Address {
		alive := &PingReply{}
		err := call("Node.Ping", n.Successor[0], Empty{}, alive)
		if err != nil {
			log.Println(err)
			if len(n.Successor) > 0 {
				n.Successor = n.Successor[1:]
			}
		}
	}

	reply := &NotifyReply{}
	err := call("Node.Notify", n.Successor[0], &NotifyArgs{Node: *n}, reply)
	if err != nil {
		fmt.Println(err)
	}
	successors := []string{n.Successor[0]}
	successors = append(successors, reply.Successors...)
	n.Successor = successors
}

func (n *Node) Notify(args *NotifyArgs, reply *NotifyReply) error {
	if equals(n.ID, args.Node.ID) {
		return nil
	}
	if n.Predecessor == nil || (len(n.Successor) > 0 && between(n.Predecessor.ID, args.Node.ID, hash(n.Successor[0]), false)) {
		n.Predecessor = &args.Node
	}
	fmt.Println("notify")
	reply.Successors = n.Successor
	return nil
}

func (n *Node) FixFingers() {}
func (n *Node) CheckPredecessor() {
	if n.Predecessor != nil {
		reply := &PingReply{}
		err := call("Node.Ping", n.Predecessor.Address, Empty{}, reply)
		if !reply.Alive || err != nil {
			n.Predecessor = nil
		}
	}
}

func (n *Node) ClosestPreceedingNode(id string) string {
	// TODO: use id param with finger table
	if len(n.Successor) > 0 {
		return n.Successor[0]
	}
	return ""
}

func (n *Node) Find(args *FindArgs, reply *FindReply) error {
	next := args.Start
	found := false
	i := 0
	for !found || i < maxRequests {
		found, next = n.FindSuccessor(args.ID)
		i++
	}
	if found {
		reply.ID = next
		return nil
	}

	return fmt.Errorf("could not find successor")
}

func (n *Node) FindSuccessor(id string) (bool, string) {
	if between(n.ID, toBigInt(id), toBigInt(n.Successor[0]), true) {
		return true, n.Successor[0]
	}
	return false, n.Successor[0]
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

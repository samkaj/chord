package chord

import (
	"crypto/sha1"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

func hash(addr string) *big.Int {
	h := sha1.New()
	h.Write([]byte(addr))
	return new(big.Int).SetBytes(h.Sum(nil))
}

type Node struct {
	ID                       *big.Int
	Address                  string
	Successor                []*Node
	Predecessor              string
	StabilizeInterval        int
	FixFingersInterval       int
	CheckPredecessorInterval int
}

func (n *Node) CreateRing() {
	n.Predecessor = ""
	n.Successor = []*Node{n} // FIXME: should call find successor on the dest node
	n.Start()
}

// Joins a ring and sets the destination to the successor of the node.
func (n *Node) JoinRing(dest string) {
	successor := &Node{
		ID:      hash(dest),
		Address: dest,
	}
	n.Predecessor = ""
	n.Successor = []*Node{successor} // FIXME: should call find successor on the dest node
	n.Start()
}

// Calls the functions on their specified intervals.
func (n *Node) StartIntervals() {
	go callOnInverval(n.CheckPredecessorInterval, n.CheckPredecessor)
	go callOnInverval(n.StabilizeInterval, n.Stabilize)
	go callOnInverval(n.FixFingersInterval, n.FixFingers)
}

// FIXME: no ops for now
func (n *Node) Stabilize() {
	if len(n.Successor) > 0 {
		x := n.Successor[0]
    if x.Address == n.Address {
      return
    }
		err := call("Node.Notify", x.Address, &NotifyArgs{Node: *n}, &Empty{})
		if err != nil {
			log.Printf("failed to notify node: %v\n", err)
		}
	}
}

func (n *Node) FixFingers() {}
func (n *Node) CheckPredecessor() {
	if n.Predecessor != "" {
		log.Println("check predecessor")
		reply := &PingReply{}
		err := call("Node.Ping", n.Predecessor, Empty{}, reply)
		if !reply.Alive || err != nil {
			n.Predecessor = ""
		}
	}
}

func (n *Node) Ping(args *Empty, reply *PingReply) error {
	log.Println("ping")
	reply.Alive = true
	return nil
}

func (n *Node) Notify(args *NotifyArgs, reply *Empty) error {
	// TODO: check from our predecessor to the node in args
	if n.Predecessor == "" {
		n.Predecessor = args.Node.Address
	}
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
  

	log.Println("Starting RPC server on ", n.Address)
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
}

type NotifyArgs struct {
	Node Node
}

// -------------------- RPC end ----------------------

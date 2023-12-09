package chord

import (
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"
)

const null = ""

type Node struct {
	ID                       string
	Address                  string
	Successors               []NodeRef
	Predecessor              NodeRef
	FingerTable              []NodeRef
	PublicKey                []byte
	Data                     map[string]string
	StabilizeInterval        int
	FixFingersInterval       int
	CheckPredecessorInterval int
	R                        int
	M                        int
	Next                     int
	TLSAddress               string
}

// Create a new node with the given address
func (node *Node) CreateNode(address string) {

	nodeRef := new(NodeRef)
	nodeRef.Address = address
	nodeRef.TLSAddress = node.TLSAddress

	file, err := os.ReadFile("./cert.pem")
	if err != nil {
		log.Fatal("Failed to read certificate file: \n Run: openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes \n To generate a cert.pem", err)
	}

	nodeRef.PublicKey = file
	node.Address = address
	node.PublicKey = file
	node.Successors[0] = *nodeRef

	node.Predecessor = *&NodeRef{TLSAddress: null, Address: null, PublicKey: []byte(null)}
	node.FingerTable = make([]NodeRef, node.M)
	node.Data = make(map[string]string)
}

func (node *Node) Start() {
	node.Next = 0
	node.StartIntervals()
	node.ServeAndListen()
}

func (node *Node) StartIntervals() {
	callOnInterval(node.StabilizeInterval, node.Stabilize)
	callOnInterval(node.FixFingersInterval, node.FixFingers)
	callOnInterval(node.CheckPredecessorInterval, node.CheckPredecessor)
}

// Join an existing ring
func (node *Node) Join(address string) {
	args := new(FindSuccessorArgs)
	args.Key = node.Address
	reply := new(FindSuccessorReply)
	log.Printf("Joining %s\n", address)
	err := call("Node.FindSuccessor", address, args, reply)
	if err != nil {
		log.Fatal(err)
	}
	node.Successors[0] = reply.Successor
	log.Printf("Successor: %s\n", node.Successors[0])
	node.Start()
}

// Find the successor of a given key
func (node *Node) FindSuccessor(args *FindSuccessorArgs, reply *FindSuccessorReply) error {
	if between(Hash(node.Address), Hash(args.Key), Hash(node.Successors[0].Address), true) {

		reply.Successor = node.Successors[0]
	} else {
		closestPrecedingNodeArgs := new(ClosestPrecedingNodeArgs)
		closestPrecedingNodeArgs.Key = args.Key
		closestPrecedingNodeReply := new(ClosestPrecedingNodeReply)
		err := call("Node.ClosestPrecedingNode", node.Address, closestPrecedingNodeArgs, closestPrecedingNodeReply)
		if err != nil {
			log.Fatal(err)
		}
		err = call("Node.FindSuccessor", closestPrecedingNodeReply.Node.Address, args, reply)
		if err != nil {
			log.Fatal(err)
		}

		reply.Successor = closestPrecedingNodeReply.Node
	}
	return nil
}

// Notify a node that it may be its predecessor
func (node *Node) Notify(args *NotifyArgs, reply *Empty) error {
	if node.Predecessor.Address == "" || between(Hash(node.Predecessor.Address), Hash(args.Key.Address), Hash(node.Address), false) {
		node.Predecessor = args.Key
	}
	return nil
}

func (node *Node) GetSuccessorList(args *GetSuccessorlistArgs, reply *GetSuccessorlistReply) error {
	reply.Successors = node.Successors
	return nil
}

func (node *Node) ClosestPrecedingNode(args *ClosestPrecedingNodeArgs, reply *ClosestPrecedingNodeReply) error {
	for i := node.M - 1; i > 0; i-- {
		if node.FingerTable[i].Address != "" && between(Hash(node.Address), Hash(node.FingerTable[i].Address), Hash(args.Key), false) {
			reply.Node = node.FingerTable[i]
			return nil
		}
	}
	reply.Node = node.Successors[0]
	return nil

}

// Stabilize the ring
func (node *Node) Stabilize() {
	x := new(GetPredecessorReply)
	x.Predecessor = node.Predecessor
	if node.Successors[0].Address != node.Address {
		x = new(GetPredecessorReply)
		call("Node.GetPredecessor", node.Successors[0].Address, &Empty{}, x)
	}

	// node ∃ (Predecessor, Successor)
	if x.Predecessor.Address != "" && between(Hash(node.Address), Hash(x.Predecessor.Address), Hash(node.Successors[0].Address), false) {
		node.Successors[0] = x.Predecessor
	}

	notifyArgs := new(NotifyArgs)
	notifyArgs.Key = *&NodeRef{Address: node.Address, PublicKey: node.PublicKey, TLSAddress: node.TLSAddress}
	notifyReply := new(NotifyReply)

	if node.Successors[0].Address == node.Address {
		return
	}

	err := call("Node.Notify", node.Successors[0].Address, notifyArgs, notifyReply)
	if err != nil {
		node.Successors = node.Successors[1:]
	}

	if len(node.Successors) == 0 {
		temp := *&NodeRef{TLSAddress: node.TLSAddress, Address: node.Address, PublicKey: node.PublicKey}
		node.Successors = append(node.Successors, temp)
	}

	getSuccessorlistArgs := new(GetSuccessorlistArgs)
	getSuccessorlistReply := new(GetSuccessorlistReply)
	err = call("Node.GetSuccessorList", node.Successors[0].Address, getSuccessorlistArgs, getSuccessorlistReply)
	if err != nil {
		return
	}
	var successorlistReply []NodeRef
	if len(getSuccessorlistReply.Successors) >= node.R {
		successorlistReply = getSuccessorlistReply.Successors[:node.R-1]
	}
	node.Successors = append([]NodeRef{node.Successors[0]}, successorlistReply...)

}

// Fix the finger table of a given node
func (node *Node) FixFingers() {
	node.Next = node.Next + 1
	if node.Next > node.M {
		node.Next = 1
	}
	log.Println("Fixing finger: ", node.Next)
	succArgs := new(FindSuccessorArgs)
	succArgs.Key = big.NewInt(2).Exp(big.NewInt(2), big.NewInt(int64(node.Next-1)), nil).String()
	succReply := new(FindSuccessorReply)
	err := call("Node.FindSuccessor", node.Address, succArgs, succReply)
	if err != nil {
		return
	}
	node.FingerTable[node.Next] = succReply.Successor
}

// Check the predecessor of a given node
func (node *Node) CheckPredecessor() {
	var successors []string
	for i, v := range node.Successors {

		successors = append(successors, fmt.Sprintf("{%d: "+v.Address+" %d"+"}", i, len(v.PublicKey)))
	}
	var fingers []string
	for i, v := range node.FingerTable {

		fingers = append(fingers, fmt.Sprintf("{%d: "+v.Address+" %d"+"}", i, len(v.PublicKey)))
	}
	err := call("Node.Ping", node.Predecessor.Address, &Empty{}, &Empty{})
	if err != nil {
		node.Predecessor = *&NodeRef{Address: null, PublicKey: []byte(null), TLSAddress: null}
	}
}

func (node *Node) GetInfo() string {
	var info strings.Builder
	info.WriteString("Node:\n")
	info.WriteString(fmt.Sprintf("  ID: %s\n  Address: %s\n\n", node.ID, node.Address))
	info.WriteString("Successors:\n")
	for _, s := range node.Successors {
		info.WriteString(fmt.Sprintf("  ID: %s\n  Address: %s\n\n", Hash(s.Address), s.Address))
	}
	info.WriteString("Fingers:\n")
	for _, finger := range node.FingerTable {
		if finger.Address != "" {
			info.WriteString(fmt.Sprintf("  ID: %s\n  Address: %s\n\n", Hash(finger.Address), finger.Address))
		}
	}
	return info.String()
}

func (node *Node) Ping(args *Empty, reply *Empty) error {
	return nil
}

func (node *Node) GetPredecessor(args *Empty, reply *GetPredecessorReply) error {
	reply.Predecessor = node.Predecessor
	return nil
}

// Calls a function on an interval
func callOnInterval(interval int, function func()) {
	go func() {
		for {
			function()
			time.Sleep(time.Duration(interval) * time.Millisecond)
		}
	}()
}

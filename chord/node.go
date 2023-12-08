package chord

import (
	"fmt"
	"log"
	"time"
)

const null = ""

type Node struct {
	ID                       string
	Address                  string
	Successor                string
	Predecessor              string
	FingerTable              []string
	Data                     map[string]string
	StabilizeInterval        int
	FixFingersInterval       int
	CheckPredecessorInterval int
}

// Create a new node with the given address
func (node *Node) CreateNode(address string) {
	node.Address = address
	node.Successor = address
	node.Predecessor = null
	node.FingerTable = make([]string, 0)
	node.Data = make(map[string]string)
}

func (node *Node) Start() {
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
	args.CallingNode = node
	reply := new(FindSuccessorReply)
	log.Printf("Joining %s\n", address)
	err := call("Node.FindSuccessor", address, args, reply)
	if err != nil {
		log.Fatal(err)
	}
	node.Successor = reply.Successor
	log.Printf("Successor: %s\n", node.Successor)
	node.Start()
}

// Find the successor of a given key
func (node *Node) FindSuccessor(args *FindSuccessorArgs, reply *FindSuccessorReply) error {
	// Prevent infinite loops
	if args.CallingNode.ID == node.ID {
		return nil
	}
	if node.Successor == node.Address {
		reply.Successor = node.Successor
		return nil
	}

	if between(ToBigInt(node.ID), ToBigInt(args.CallingNode.ID), Hash(node.Successor), true) {
		reply.Successor = node.Successor
	} else {
		closestPrecedingNodeArgs := new(ClosestPrecedingNodeArgs)
		closestPrecedingNodeArgs.CallingNode = args.CallingNode
		closestPrecedingNodeReply := new(ClosestPrecedingNodeReply)
		err := call("Node.ClosestPrecedingNode", node.Address, closestPrecedingNodeArgs, closestPrecedingNodeReply)
		if err != nil {
			log.Fatal(err)
		}
		err = call("Node.FindSuccessor", closestPrecedingNodeReply.Node, args, reply)
		if err != nil {
			log.Fatal(err)
		}
		reply.Successor = closestPrecedingNodeReply.Node
	}
	return nil
}

// Notify a node that it may be its predecessor
func (node *Node) Notify(args *NotifyArgs, reply *Empty) error {
	if node.Predecessor == "" || between(Hash(node.Predecessor), ToBigInt(args.CallingNode.ID), ToBigInt(node.ID), false) {
		fmt.Printf("Setting predecessor to: %s\n", args.CallingNode.Address)
		node.Predecessor = args.CallingNode.Address
	}
	return nil
}

func (node *Node) ClosestPrecedingNode(args *ClosestPrecedingNodeArgs, reply *ClosestPrecedingNodeReply) error {
	// TODO: use finger table
	reply.Node = node.Address
	return nil

}

// Update the finger table of a given node
func (node *Node) UpdateFingerTable(key string, s int) {
	// TODO
	log.Fatal("Not implemented")
}

// Update the successor of a given node
func (node *Node) UpdateSuccessor() {
	// TODO
	log.Fatal("Not implemented")
}

// Update the predecessor of a given node
func (node *Node) UpdatePredecessor() {
	// TODO
	log.Fatal("Not implemented")
}

// Stabilize the ring
func (node *Node) Stabilize() {

	x := new(GetPredecessorReply)
	x.Predecessor = node.Predecessor
	if node.Successor != node.Address {
		x = new(GetPredecessorReply)
		call("Node.GetPredecessor", node.Successor, &Empty{}, x)
		fmt.Println("GetPredecessorReply: ", x)
	}

	// node ∃ (Predecessor, Successor)
	if x.Predecessor != "" && between(Hash(node.Address), Hash(x.Predecessor), Hash(node.Successor), false) {
		fmt.Printf("Setting successor \n")
		node.Successor = x.Predecessor
	}

	notifyArgs := new(NotifyArgs)
	notifyArgs.CallingNode = node
	notifyReply := new(NotifyReply)
	// You are your own successor

	if node.Successor == node.Address {
		return
	}

	err := call("Node.Notify", node.Successor, notifyArgs, notifyReply)
	if err != nil {
		log.Fatal(err)
	}

}

// Fix the finger table of a given node
func (node *Node) FixFingers() {
	// TODO
}

// Check the predecessor of a given node
func (node *Node) CheckPredecessor() {
	// TODO
	fmt.Println("--------Node--------")
	fmt.Println("ID: ", node.ID)
	fmt.Println("Adress: ", node.Address)
	fmt.Println("Successor: ", node.Successor)
	fmt.Println("Predecessor: ", node.Predecessor)
	fmt.Println("--------------------")
	err := call("Node.Ping", node.Predecessor, &Empty{}, &Empty{})
	if err != nil {
		node.Predecessor = ""
	}
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

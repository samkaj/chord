package chord

import (
	"log"
	"time"
)

const null = ""

type Node struct {
	ID                       string
	Address                  string
	Successor                []string
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
	node.Successor = []string{address}
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
	if args.CallingNode.Address == node.Address {
		reply.Successor = node.Successor
		return nil
	}

	if between(ToBigInt(node.ID), ToBigInt(args.CallingNode.ID), Hash(node.Successor[0]), true) {
		reply.Successor = node.Successor
	} else {
		closestPrecedingNodeArgs := new(ClosestPrecedingNodeArgs)
		closestPrecedingNodeArgs.CallingNode = args.CallingNode
		closestPrecedingNodeReply := new(ClosestPrecedingNodeReply)
		err := call("Node.ClosestPrecedingNode", node.Address, closestPrecedingNodeArgs, closestPrecedingNodeReply)
		if err != nil {
			log.Fatal(err)
		}
		reply.Successor = []string{closestPrecedingNodeReply.Node}
	}
	return nil
}

// Notify a node that it may be its predecessor
func (node *Node) Notify(args *NotifyArgs, reply *Empty) error {
	if node.Address != args.CallingNode.Address && (node.Predecessor == "" || between(Hash(node.Predecessor), ToBigInt(args.CallingNode.ID), ToBigInt(node.ID), false)) {
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
	if node.Successor[0] == node.ID {
		return
	}

	x := new(GetPredecessorReply)
	call("Node.GetPredecessor", node.Successor[0], &Empty{}, x)

	if x.Predecessor != null && between(Hash(x.Predecessor), ToBigInt(node.ID), Hash(node.Successor[0]), false) {
		// node.Successor = x.Predecessor
    node.Successor = append(node.Successor, x.Predecessor)
	}

	notifyArgs := new(NotifyArgs)
	notifyArgs.CallingNode = node
	notifyReply := new(NotifyReply)
  log.Println(node.Successor)
	err := call("Node.Notify", node.Successor[0], notifyArgs, notifyReply)
	if err != nil {
		log.Fatal(err)
	}
}

func (node *Node) Ping(args *Empty, reply *PingReply) error {
	reply.Alive = true
	return nil
}

// Fix the finger table of a given node
func (node *Node) FixFingers() {
	// TODO
}

// Check the predecessor of a given node
func (node *Node) CheckPredecessor() {
	reply := new(PingReply)
	err := call("Node.Ping", node.Predecessor, &Empty{}, reply)
	if err != nil || !reply.Alive {
		node.Predecessor = null
	}
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

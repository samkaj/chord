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
	Successor                *Node
	Predecessor              *Node
	FingerTable              []string
	Data                     map[string]string
	StabilizeInterval        int
	FixFingersInterval       int
	CheckPredecessorInterval int
}

// Create a new node with the given address
func (node *Node) CreateNode(address string) {
	node.Address = address
	node.Successor = node
	node.Predecessor = nil
	node.FingerTable = make([]string, 0)
	node.Data = make(map[string]string)
}

func (node *Node) Start() {
  node.StartIntervals()
	node.ServeAndListen()
}

func (node *Node) StartIntervals () {
	callOnInterval(node.StabilizeInterval, node.Stabilize)
	callOnInterval(node.FixFingersInterval, node.FixFingers)
	callOnInterval(node.CheckPredecessorInterval, node.CheckPredecessor)
}

// Join an existing ring
func (node *Node) Join(address string) {
  args := new(FindSuccessorArgs) 
  args.CallingNode = node
  //reply := new(FindSuccessorReply)
  log.Printf("Joining %s\n", address)
  log.Fatal("not implemented")
  // whut the heeeeeeeeeeeeeeeeell
  //err := call("FindSuccessor", address, args, reply)
  //if err != nil {
  //  log.Fatal(err)
  //}
  //node.Successor = reply.Successor
  //log.Printf("Successor: %s\n", node.Successor.Address)
  //log.Fatal("not implemented")
  //node.Start()
}

// Find the successor of a given key
func (node *Node) FindSuccessor(args *FindSuccessorArgs, reply *FindSuccessorReply) error {
  // Prevent infinite loops 
  if args.CallingNode.ID == node.ID {
    return nil
  }

  log.Printf("FindSuccessor: %s\n", args.CallingNode.Address)

  
  if between(ToBigInt(node.ID), ToBigInt(args.CallingNode.ID), ToBigInt(node.Successor.ID), true) {
    reply.Successor = node.Successor
  } else {
    // forward the query around the circle
    closestPrecedingNodeArgs := new(ClosestPrecedingNodeArgs)
    closestPrecedingNodeArgs.CallingNode = args.CallingNode
    closestPrecedingNodeReply := new(ClosestPrecedingNodeReply)
    err := call("ClosestPrecedingNode", node.Address, closestPrecedingNodeArgs, closestPrecedingNodeReply)
    if err != nil {
      log.Fatal(err)
    }
    err = call("FindSuccessor", closestPrecedingNodeReply.Node.Address, args, reply)
    if err != nil {
      log.Fatal(err)
    }
    reply.Successor = closestPrecedingNodeReply.Node
  }
	return nil
}

// Notify a node that it may be its predecessor
func (node *Node) Notify(args *NotifyArgs, reply *Empty) error {
  if node.Predecessor == nil || between(ToBigInt(node.Predecessor.ID), ToBigInt(args.CallingNode.ID), ToBigInt(node.ID), false) {
    node.Predecessor = args.CallingNode
  }
  return nil
}

func (node *Node) ClosestPrecedingNode(args *ClosestPrecedingNodeArgs, reply *ClosestPrecedingNodeReply) error {
  // TODO: use finger table
  reply.Node = node
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
  if node.Successor.ID == node.ID {
    return
  }

  x := node.Successor.Predecessor
  if x != nil && between(ToBigInt(x.ID), ToBigInt(node.ID), ToBigInt(node.Successor.ID), false) {
    node.Successor = x
  }

  notifyArgs := new(NotifyArgs)
  notifyArgs.CallingNode = node
  notifyReply := new(NotifyReply)
  err := call("Notify", node.Successor.Address, notifyArgs, notifyReply)
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
  fmt.Println(node.Predecessor)
}

func (node *Node) GetPredecessor(args *Empty, reply *FindSuccessorReply) error {
  reply.Successor = node.Predecessor
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

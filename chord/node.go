package chord

import (
	"log"
	"math/big"
	"os"
	"time"
)

const null = ""

type Node struct {
	ID                       string
	Address                  string
	Successors               []NodeRef
	Predecessor              NodeRef
	FingerTable              []NodeRef
	PublicKey                string
	Data                     map[string]string
	StabilizeInterval        int
	FixFingersInterval       int
	CheckPredecessorInterval int
	R                        int
	M                        int
	Next                     int
}

// Create a new node with the given address
func (node *Node) CreateNode(address string) {

	nodeRef := new(NodeRef)
	nodeRef.Address = address

	file, err := os.ReadFile("./cert.pem")
	if err != nil {
		log.Fatal("Failed to read certificate file: \n Run: openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes \n To generate a cert.pem", err)
	}
	nodeRef.publicKey = string(file)
	node.Address = address
	node.Successors[0] = *nodeRef
	node.Predecessor = *&NodeRef{Address: null, publicKey: null}
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
		log.Printf("Setting predecessor to: %s\n", args.Key)
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
	if node.Successors[0].Address != node.Address {
		x = new(GetPredecessorReply)
		call("Node.GetPredecessor", node.Successors[0].Address, &Empty{}, x)
		log.Println("GetPredecessorReply: ", x)
	}

	// node ∃ (Predecessor, Successor)
	if x.Predecessor.Address != "" && between(Hash(node.Address), Hash(x.Predecessor.Address), Hash(node.Successors[0].Address), false) {
		log.Printf("Setting successor \n")
		node.Successors[0] = x.Predecessor
	}

	notifyArgs := new(NotifyArgs)
	notifyArgs.Key = *&NodeRef{Address: node.Address, publicKey: node.PublicKey}
	notifyReply := new(NotifyReply)

	if node.Successors[0].Address == node.Address {
		return
	}

	err := call("Node.Notify", node.Successors[0].Address, notifyArgs, notifyReply)
	if err != nil {
		node.Successors = node.Successors[1:]
	}

	if len(node.Successors) == 0 {
		temp := *&NodeRef{Address: node.Address, publicKey: node.PublicKey}
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
	// TODO
	log.Println("--------Node--------")
	log.Println("ID: ", node.ID)
	log.Println("Adress: ", node.Address)
	log.Println("Successors: ", node.Successors)
	log.Println("Predecessor: ", node.Predecessor)
	log.Println("--------------------")
	err := call("Node.Ping", node.Predecessor.Address, &Empty{}, &Empty{})
	if err != nil {
		node.Predecessor = *&NodeRef{Address: null, publicKey: null}
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

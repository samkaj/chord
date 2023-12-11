package chord

import (
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"
)

const redundancy = 3 // Number of copies by hashing the path to send it to different nodes.

type Node struct {
	ID                       string    // ID is the hash of the address
	Address                  string    // Address is the IP address of the node
	Successors               []NodeRef // Successors is a list of successors
	Predecessor              NodeRef   // Predecessor is the predecessor of the node
	FingerTable              []NodeRef // FingerTable is the finger table of the node
	PublicKey                []byte    // PublicKey is the public key of the node. Used for TLS
	StabilizeInterval        int       // StabilizeInterval is the interval at which the node stabilizes
	FixFingersInterval       int       // FixFingersInterval is the interval at which the node fixes its finger table
	CheckPredecessorInterval int       // CheckPredecessorInterval is the interval at which the node checks its predecessor
	R                        int       // R is the number of successors to keep in the successor list
	M                        int       // M is the number of entries in the finger table, matches the identifier space
	Next                     int       // Next is the next finger to fix
	TLSAddress               string    // TLSAddress is the address to listen for TLS connections on
	StoragePath              string    // StoragePath is the path to the storage directory
}

// Create a new node with the given address
func (node *Node) CreateNode() {
	nodeRef := new(NodeRef)
	nodeRef.TLSAddress = node.TLSAddress

	file, err := os.ReadFile("./cert.pem")
	if err != nil {
		log.Fatal("Failed to read certificate file: \n Run: openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes \n To generate a cert.pem", err)
	}

	nodeRef.PublicKey = file
	node.PublicKey = file
	node.Successors[0] = *nodeRef
	node.Predecessor = NodeRef{TLSAddress: "", Address: "", PublicKey: []byte("")}
	node.FingerTable = make([]NodeRef, node.M)
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
	node.Start()
}

func bytesToBigInt(b []byte) *big.Int {
	return new(big.Int).SetBytes(b)
}

// Find the successor of a given key
func (node *Node) FindSuccessor(args *FindSuccessorArgs, reply *FindSuccessorReply) error {
	num := new(big.Int)
	num.SetString(args.Key, 10)
	if between(Hash(node.Address), num, Hash(node.Successors[0].Address), true) {
		reply.Successor = node.Successors[0]
	} else {
		closestPrecedingNodeArgs := new(ClosestPrecedingNodeArgs)
		closestPrecedingNodeArgs.Key = num.String()
		closestPrecedingNodeReply := new(ClosestPrecedingNodeReply)
		err := call("Node.ClosestPrecedingNode", node.Address, closestPrecedingNodeArgs, closestPrecedingNodeReply)
		if err != nil {
			return err
		}

		err = call("Node.FindSuccessor", closestPrecedingNodeReply.Node.Address, args, reply)
		if err != nil {
			return err
		}
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

// Used to let someone inherit their successors successor list
func (node *Node) GetSuccessorList(args *GetSuccessorlistArgs, reply *GetSuccessorlistReply) error {
	reply.Successors = node.Successors
	return nil
}

// Get highest predecessor to a node in finger table
func (node *Node) ClosestPrecedingNode(args *ClosestPrecedingNodeArgs, reply *ClosestPrecedingNodeReply) error {
	num := new(big.Int)
	num.SetString(args.Key, 10)
	for i := node.M - 1; i > 0; i-- {
		if node.FingerTable[i].Address != "" && between(Hash(node.Address), Hash(node.FingerTable[i].Address), num, false) {
			reply.Node = node.FingerTable[i]
			return nil
		}
	}
	reply.Node = NodeRef{Address: node.Address, PublicKey: node.PublicKey, TLSAddress: node.TLSAddress}
	return nil
}

// Stores a file in the ring by finding the correct succesor and then using TLSSend to send the file to the successor.
// We use redundancy to store the file on multiple nodes. This is done by hashing multiple times.
func (node *Node) Store(path string, data []byte) error {
	hashedPath := path
	for i := 0; i < redundancy; i++ {
		succArgs := new(FindSuccessorArgs)
		if i == 0 {
			hashedPath = path
		} else {
			for j := 0; j < i; j++ {
				hashedPath = Hash(hashedPath).String()
			}
		}
		succArgs.Key = hashedPath
		succReply := new(FindSuccessorReply)
		err := call("Node.FindSuccessor", node.Address, succArgs, succReply)
		if err != nil {
			return fmt.Errorf("failed to find successor: %w", err)
		}
		TLSSend(succReply.Successor, path, data)
	}
	return nil
}

// Get a file from the ring. Since we are using redundancy, we can just get the file from the first node that has it.
func (node *Node) GetFile(path string) ([]byte, error) {
	for i := 0; i < redundancy; i++ {
		succArgs := new(FindSuccessorArgs)
		if i == 0 {
			succArgs.Key = path
		} else {
			succArgs.Key = Hash(path).String()
		}
		succReply := new(FindSuccessorReply)
		err := call("Node.FindSuccessor", node.Address, succArgs, succReply)
		if err != nil {
			return nil, fmt.Errorf("failed to find successor: %w", err)
		}
		data, err := TLSGet(succReply.Successor, path)
		if err == nil {
			return data, nil
		}
	}
	return nil, fmt.Errorf("failed to get file")
}

// Verifies the immediate successor and tells the successor about this node
func (node *Node) Stabilize() {
	// Get predecessor of our successor
	x := new(GetPredecessorReply)
	x.Predecessor = node.Predecessor
	if node.Successors[0].Address != node.Address {
		x = new(GetPredecessorReply)
		call("Node.GetPredecessor", node.Successors[0].Address, &Empty{}, x)
	}

	// If x is between this node and its successor, set successor to x
	if x.Predecessor.Address != "" && between(Hash(node.Address), Hash(x.Predecessor.Address), Hash(node.Successors[0].Address), false) {
		node.Successors[0] = x.Predecessor
	}

	// Ad-hoc fix for when the successor list is empty
	if node.Successors[0].Address == node.Address {
		return
	}

	notifyArgs := new(NotifyArgs)
	notifyArgs.Key = *&NodeRef{Address: node.Address, PublicKey: node.PublicKey, TLSAddress: node.TLSAddress}
	notifyReply := new(NotifyReply)
	err := call("Node.Notify", node.Successors[0].Address, notifyArgs, notifyReply)
	if err != nil {
		// If the successor is down, remove it from the successor list
		node.Successors = node.Successors[1:]
	}

	// Set ourselves as successor if the successor list is empty
	if len(node.Successors) == 0 {
		temp := *&NodeRef{TLSAddress: node.TLSAddress, Address: node.Address, PublicKey: node.PublicKey}
		node.Successors = append(node.Successors, temp)
	}

	// Get successors from our successor
	getSuccessorlistArgs := new(GetSuccessorlistArgs)
	getSuccessorlistReply := new(GetSuccessorlistReply)
	err = call("Node.GetSuccessorList", node.Successors[0].Address, getSuccessorlistArgs, getSuccessorlistReply)
	if err != nil {
		return
	}

	successorlistReply := getSuccessorlistReply.Successors
	if len(getSuccessorlistReply.Successors) >= node.R {
		successorlistReply = successorlistReply[:node.R-1]
	}

	// Append our successor to the successor list
	node.Successors = append([]NodeRef{node.Successors[0]}, successorlistReply...)
}

// Fix the finger table of a given node
func (node *Node) FixFingers() {
	node.Next = (node.Next + 1%node.M)
	if node.Next >= node.M {
		// Stay in bounds
		node.Next = 1
	}
	succArgs := new(FindSuccessorArgs)

	// From paper: n + 2^(next-1)
	bigN := Hash(node.Address)
	two := big.NewInt(2)
	exponent := big.NewInt(int64(node.Next - 1))
	twoToThePower := new(big.Int).Exp(two, exponent, nil)
	x := new(big.Int).Add(bigN, twoToThePower)
	succArgs.Key = x.String()
	succReply := new(FindSuccessorReply)
	err := call("Node.FindSuccessor", node.Address, succArgs, succReply)
	if err != nil {
		return
	}
	node.FingerTable[node.Next] = succReply.Successor
}

// Check the predecessor of a given node
func (node *Node) CheckPredecessor() {
	err := call("Node.Ping", node.Predecessor.Address, &Empty{}, &Empty{})
	if err != nil {
		node.Predecessor = NodeRef{}
	}
}

// Prints
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

// Ping function for checking if a node is alive
func (node *Node) Ping(args *Empty, reply *Empty) error {
	return nil
}

// Get the predecessor of a node
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

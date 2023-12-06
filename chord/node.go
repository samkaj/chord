package chord

import (
	"log"
	"time"
  "math/big"
)

const null = ""

type Node struct {
	ID                       *big.Int
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
	log.Printf("Creating new ring with address %s", address)
	node.Address = address
	node.Successor = address
	node.Predecessor = null
	node.FingerTable = make([]string, 0)
	node.Data = make(map[string]string)
}

func (node *Node) Start() {
	callOnInterval(node.StabilizeInterval, node.Stabilize)
	callOnInterval(node.FixFingersInterval, node.FixFingers)
	callOnInterval(node.CheckPredecessorInterval, node.CheckPredecessor)
	node.ServeAndListen()
}

// Join an existing ring
func (node *Node) Join(address string) {
	// TODO
}

// Find the successor of a given key
func (node *Node) FindSuccessor(key string) string {
	// TODO
	return null
}

// Find the predecessor of a given key
func (node *Node) FindPredecessor(key string) string {
	// TODO
	return null
}

// Find the closest preceding finger of a given key
func (node *Node) ClosestPrecedingFinger(key string) string {
	// TODO
	return null
}

// Update the finger table of a given node
func (node *Node) UpdateFingerTable(key string, s int) {
	// TODO
}

// Update the successor of a given node
func (node *Node) UpdateSuccessor() {
	// TODO
}

// Update the predecessor of a given node
func (node *Node) UpdatePredecessor() {
	// TODO
}

// Stabilize the ring
func (node *Node) Stabilize() {
	// TODO
}

// Fix the finger table of a given node
func (node *Node) FixFingers() {
	// TODO
}

// Check the predecessor of a given node
func (node *Node) CheckPredecessor() {
	// TODO
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

package chord

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type CLI struct {
	Node *Node
}

// Reads from stdin and handles commands.
func (c *CLI) ReadCommands(node *Node) {
	scanner := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprintf(os.Stdout, "\033[1;34m♫\033[0m ")
		command, _ := scanner.ReadString('\n')
		c.handleCommand(command)
	}
}

// Handles a command from the user in the CLI.
func (c *CLI) handleCommand(args string) {
	if args == "\n" {
		return
	}
	args = strings.TrimSpace(args)
	parts := strings.Split(args, " ")
	command := parts[0]
	param := ""
	if len(parts) > 1 {
		param = parts[1]
	}
	switch command {
	case "lookup":
		c.lookup(param)
	case "store":
		c.storeFile(param)
	case "print":
		c.printState()
	case "exit":
		c.exit()
	case "clear":
		c.clear()
	case "help":
		c.usage()
	case "test":
		c.test(param)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		c.usage()
	}
}

// Given a path to a file, the client takes this file, hashes it to a key in the identifier space,
// and performs a search for the node that is the successor to the file.
func (c *CLI) lookup(key string) {
	if key == "" {
		fmt.Fprintf(os.Stderr, "No key supplied\n")
		return
	}

	fmt.Fprintf(os.Stdout, c.findFile(key))
}

// Finds the successor with a given key and returns the file information.
func (c *CLI) findFile(key string) string {
	reply := new(FindSuccessorReply)
	args := new(FindSuccessorArgs)
	args.Key = Hash(key).String()

	err := call("Node.FindSuccessor", c.Node.Address, args, reply)
	if err != nil {
		return "Failed to find successor\n"
	}

	addr := reply.Successor.Address
	data, err := TLSGet(reply.Successor, key)
	if err != nil {
		return "Failed to get file\n"
	}
	return fmt.Sprintf("ID: %s\nAddress: %s\nContent:\n%s\n", Hash(addr), addr, data)
}

// Useful test method for finding the successor of a given key
func (c *CLI) test(key string) {
	if key == "" {
		fmt.Fprintf(os.Stderr, "No key supplied\n")
		return
	}
	reply := new(FindSuccessorReply)
	args := new(FindSuccessorArgs)
	args.Key = Hash(key).String()

	err := call("Node.FindSuccessor", c.Node.Address, args, reply)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to find successor\n")
		return
	}
	fmt.Fprintf(os.Stdout, "FileHash %v \nID: %v \nAdress: %v \n", Hash(key), Hash(reply.Successor.Address), reply.Successor.Address)
}

// Takes the location of a file on a local disk, then performs a lookup.
// Once the correct place of the file is found, the file gets uploaded to the Chord ring.
func (c *CLI) storeFile(path string) {
	if path == "" {
		fmt.Fprintf(os.Stderr, "No path supplied\n")
		return
	}
	data, err := readFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read file: %s\n", err)
		return
	}
	c.Node.Store(path, data)
}

// Outputs its local state information at the current time, which consists of:
// 1. The client's own node information
// 2. The node information for all nodes in the successor list
// 3. The node information for all nodes in the finger table
// where “node information” corresponds to the identifier, IP address, and port for a given node.
func (c *CLI) printState() {
	fmt.Fprintf(os.Stdout, "%s\n", c.Node.GetInfo())
}

// Prints the usage message.
func (c *CLI) usage() {
	usage := `Usage: [command]
Commands:
  lookup [key] - lookup a file with the given key
  store [path] - store a file with the given path
  print        - print the state of the client
  exit         - exit the client
  help         - print this message
`
	fmt.Fprintf(os.Stdout, usage)
}

// Clears the screen.
func (c *CLI) clear() {
	fmt.Fprintf(os.Stdout, "\033[2J\033[1;1H")
}

// Exits the client.
func (c *CLI) exit() {
	fmt.Fprintf(os.Stdout, "Exiting client...\n")
	os.Exit(0)
}

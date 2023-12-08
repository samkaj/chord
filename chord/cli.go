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

func (c *CLI) ReadCommands(node *Node) {
	scanner := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprintf(os.Stdout, "\033[1;34m♫\033[0m ")
		command, _ := scanner.ReadString('\n')
		c.handleCommand(command)
	}
}

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
	fmt.Fprintf(os.Stdout, "Lookup(%s)\n", key)
}

// Takes the location of a file on a local disk, then performs a lookup.
// Once the correct place of the file is found, the file gets uploaded to the Chord ring.
func (c *CLI) storeFile(path string) {
	if path == "" {
		fmt.Fprintf(os.Stderr, "No path supplied\n")
		return
	}
	fmt.Fprintf(os.Stdout, "StoreFile(%s)\n", path)
}

// Outputs its local state information at the current time, which consists of:
// 1. The client's own node information
// 2. The node information for all nodes in the successor list
// 3. The node information for all nodes in the finger table
// where “node information” corresponds to the identifier, IP address, and port for a given node.
func (c *CLI) printState() {
	fmt.Fprintf(os.Stdout, "PrintState()\n")
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

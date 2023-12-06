package main

import (
	"chord/chord"
	"flag"
	"fmt"
	"log"
	"time"
)

func main() {
	fmt.Println("welcome to chord from wish")
	join := flag.String("j", "", "join address")
  addr := flag.String("a", "", "chord address")
	flag.Parse()

	node := new(chord.Node)
	node.CheckPredecessorInterval = 1000
  node.StabilizeInterval = 2000
	node.Address = *addr
	if *join != "" {
		log.Println("joining a ring")
		node.JoinRing(*join)
	} else {
		log.Println("creating a new ring")
		node.CreateRing()
	}

	for {
		time.Sleep(time.Second)
	}
}

package main

import (
	"chord/chord"
	"flag"
	"log"
	"time"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	a := flag.String("a", "", "the chord address")
	j := flag.String("j", "", "the join address")
	tcp := flag.Int("tcp", 0, "check predecessor interval")
	ts := flag.Int("ts", 0, "stabilize interval")
	tff := flag.Int("ff", 0, "fix fingers interval")
	flag.Parse()
	node := chord.Node{}
	node.CheckPredecessorInterval = *tcp
	node.StabilizeInterval = *ts
	node.FixFingersInterval = *tff
	node.ID = chord.Hash(*a).String()
	log.Println("addr", *a)
	node.CreateNode(*a)
	if *j == "" {
		node.Start()
	}
	if *j != "" {
		node.Join(*j)
	}

	for {
		time.Sleep(time.Second)
	}
}

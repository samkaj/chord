package main

import (
	"chord/chord"
	"flag"
	"log"
	"os"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	logFile := "log.txt"
	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	a := flag.String("a", "", "the chord address")
	j := flag.String("j", "", "the join address")
	tcp := flag.Int("tcp", 0, "check predecessor interval")
	ts := flag.Int("ts", 0, "stabilize interval")
	tff := flag.Int("ff", 0, "fix fingers interval")
	r := flag.Int("r", 1, "number of successors maintained")
	flag.Parse()
	node := chord.Node{}

	node.M = 160
	node.CheckPredecessorInterval = *tcp
	node.StabilizeInterval = *ts
	node.FixFingersInterval = *tff
	node.Successors = make([]chord.NodeRef, *r)
	node.R = *r
	node.ID = chord.Hash(*a).String()
	log.Println("addr", *a)
	node.CreateNode(*a)
	if *j == "" {
		go node.Start()
	}
	if *j != "" {
		go node.Join(*j)
	}

	cli := chord.CLI{Node: &node}
	cli.ReadCommands(&node)
}

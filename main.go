package main

import (
	"chord/chord"
	"flag"
	"fmt"
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
	p := flag.Int("p", 0, "the chord port")
	ja := flag.String("ja", "", "the join address")
	jp := flag.Int("jp", 0, "the join port")
	tcp := flag.Int("tcp", 0, "check predecessor interval")
	ts := flag.Int("ts", 0, "stabilize interval")
	tff := flag.Int("ff", 0, "fix fingers interval")
	r := flag.Int("r", 0, "number of successors maintained")
	tls := flag.String("tls", "", "the tls address and port")
	flag.Parse()

	// crash if any of the required flags are not set
	if *a == "" || *p == 0 || *tcp == 0 || *ts == 0 || *tff == 0 || *r == 0 || *tls == "" || (*ja != "" && *jp == 0) || (*ja == "" && *jp != 0) {
		flag.PrintDefaults()
		os.Exit(1)
	}

	node := chord.Node{}
	log.Println("foo", *p)
	node.Address = fmt.Sprintf("%s:%d", *a, *p)
	log.Println("bar", node.Address)
	node.M = 160
	node.CheckPredecessorInterval = *tcp
	node.StabilizeInterval = *ts
	node.FixFingersInterval = *tff
	node.Successors = make([]chord.NodeRef, *r)
	node.R = *r
	node.ID = chord.Hash(*a).String()
	node.TLSAddress = *tls
	node.StoragePath = "storage-" + chord.Hash(*a).String()
	err = os.Mkdir(node.StoragePath, 0755)
	if err != nil {
		log.Println("Failed to create storage directory: ", err)
	}

	go node.TLSListen()
	node.CreateNode()
	if *jp != 0 && *ja != "" {
		j := fmt.Sprintf("%s:%d", *ja, *jp)
		go node.Join(j)
	} else {

		go node.Start()
	}

	cli := chord.CLI{Node: &node}
	cli.ReadCommands(&node)
}

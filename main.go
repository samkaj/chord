package main

import (
	"chord/chord"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
)

var ipv4Regex = regexp.MustCompile(`(^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$)|((::0)|(localhost))`)

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
	tls := flag.Int("tls", 0, "the tls port")
	flag.Parse()

	// crash if any of the required flags are not set
	if *a == "" || *p == 0 || *tcp == 0 || *ts == 0 || *tff == 0 || *r == 0 || *tls == 0 || (*ja != "" && *jp == 0) || (*ja == "" && *jp != 0) {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *ts < 1 || *ts > 60000 || *tff < 1 || *tff > 60000 || *tcp < 1 || *tcp > 60000 {
		fmt.Println("intervals should be between 1 and 60000")
		os.Exit(1)
	}

	if *r < 1 || *r > 32 {
		fmt.Println("-r should be between 1 and 32")
		os.Exit(1)
	}

	if !ipv4Regex.MatchString(*a) {
		fmt.Println("-a should be a valid IPv4 address")
		os.Exit(1)
	}

	if !ipv4Regex.MatchString(*ja) && *ja != "" {
		fmt.Println("-ja should be a valid IPv4 address")
		os.Exit(1)
	}

	node := chord.Node{}
	node.Address = fmt.Sprintf("%s:%d", *a, *p)
	node.M = 160
	node.CheckPredecessorInterval = *tcp
	node.StabilizeInterval = *ts
	node.FixFingersInterval = *tff
	node.Successors = make([]chord.NodeRef, *r)
	node.R = *r
	node.ID = chord.Hash(*&node.Address).String()
	node.TLSAddress = fmt.Sprintf("0.0.0.0:%d", *tls)
	node.StoragePath = "storage-" + chord.Hash(*&node.Address).String()
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

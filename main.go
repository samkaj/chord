package main

import (
	"chord/chord"
	"flag"
)

func main() {
	a := flag.String("a", "", "the chord address")
	j := flag.String("j", "", "the join address")
  tcp := flag.Int("tcp", 0, "check predecessor interval")
  ts := flag.Int("ts", 0, "stabilize interval")
  tff := flag.Int("ff", 0, "fix fingers interval")
	flag.Parse()
	node := chord.Node{}
	node.CreateNode(*a)
  node.CheckPredecessorInterval = *tcp
  node.StabilizeInterval = *ts
  node.FixFingersInterval = *tff
  node.ID = chord.Hash(*a)
	if *j != "" {
		node.Join(*j)
	}
}

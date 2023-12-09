package chord

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"strings"
)

func TLSListen() {
	cer, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Loaded TLS keypair: ")
	config := &tls.Config{Certificates: []tls.Certificate{cer}}

	ln, err := tls.Listen("tcp", ":3001", config)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	fmt.Println("TLS Listening on port 3001")
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		fmt.Println("Accepted TLS connection")
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Println(err)
		return
	}

	data := buffer[:n]
	fmt.Printf("Received: %s\n", string(data))
}

func TLSSend(nodeRef NodeRef, message []byte) {
	cer, err := tls.LoadX509KeyPair("cert.pem", "key.pem")

	addrSplit := strings.Split(nodeRef.Address, ":")
	if len(addrSplit) != 2 {
		log.Fatal("Invalid address")
	}

	addr := addrSplit[0] + ":3001"
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Loaded TLS keypair: ")
	config := &tls.Config{Certificates: []tls.Certificate{cer}, InsecureSkipVerify: true}

	conn, err := tls.Dial("tcp", addr, config)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	fmt.Println("TLS Connected to localhost:3001")
	conn.Write([]byte("Hello from client"))
}

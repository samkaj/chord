package chord

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"strings"
)

func (node *Node) TLSListen() {
	cer, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Loaded TLS keypair: ")
	config := &tls.Config{Certificates: []tls.Certificate{cer}}

	addrSplit := strings.Split(node.Address, ":")
	addr := addrSplit[0] + ":3001"
	ln, err := tls.Listen("tcp", addr, config)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	fmt.Println("TLS Listening on", addr)
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

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(nodeRef.PublicKey)
	config := &tls.Config{Certificates: []tls.Certificate{cer}, RootCAs: caCertPool}

	conn, err := tls.Dial("tcp", addr, config)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	fmt.Println("TLS Connected to localhost:3001")
	conn.Write([]byte("Hello from client"))
}

package chord

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
)

func (node *Node) TLSListen() {
	cer, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Loaded TLS keypair: ")
	config := &tls.Config{Certificates: []tls.Certificate{cer}}

	ln, err := tls.Listen("tcp", node.TLSAddress, config)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	log.Println("TLS Listening on", node.TLSAddress)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println("Accepted TLS connection")
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
	log.Printf("Received: %s\n", string(data))
}

func TLSSend(nodeRef NodeRef, message []byte) {
	cer, err := tls.LoadX509KeyPair("cert.pem", "key.pem")

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(nodeRef.PublicKey)
	config := &tls.Config{Certificates: []tls.Certificate{cer}, RootCAs: caCertPool}

	conn, err := tls.Dial("tcp", nodeRef.TLSAddress, config)
	if err != nil {
    fmt.Println("TLS Dial error: ", err)
		log.Fatal(err)
	}
	defer conn.Close()
	log.Println("TLS Connected to localhost:3001")
	conn.Write([]byte("Hello from client"))
}

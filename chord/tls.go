package chord

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
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

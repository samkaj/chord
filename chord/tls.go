package chord

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

// Establishes a secure channel for sending files between nodes using TLS.
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
		go node.handleConnection(conn)
	}
}

// Reads from the connection and writes the file to the node's storage.
func (node *Node) handleConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Println(err)
		return
	}

	data := buffer[:n]
	fileName := string(data[:strings.Index(string(data), "\n")])
	data = data[strings.Index(string(data), "\n")+1:]
	file, err := os.Create(node.StoragePath + "/" + fileName)
	if err != nil {
		fmt.Println("Failed to create file: ", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		fmt.Println("Failed to write to file: ", err)
	}
}

// Sends data file data to a node using TLS.
func TLSSend(nodeRef NodeRef, fileName string, data []byte) {
	cer, err := tls.LoadX509KeyPair("cert.pem", "key.pem")

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(nodeRef.PublicKey)
	config := &tls.Config{Certificates: []tls.Certificate{cer}, RootCAs: caCertPool}

	conn, err := tls.Dial("tcp", nodeRef.TLSAddress, config)
	if err != nil {
		fmt.Println("TLS Dial error: ", err)
		return
	}
	defer conn.Close()

	header := []byte(fileName + "\n")
	data = append(header, data...)
	_, err = conn.Write(data)
	if err != nil {
		fmt.Println("TLS Write error: ", err)
	}
}

// Gets a file from a node using TLS.
func TLSGet(nodeRef NodeRef, fileName string) ([]byte, error) {
	cer, err := tls.LoadX509KeyPair("cert.pem", "key.pem")

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(nodeRef.PublicKey)
	config := &tls.Config{Certificates: []tls.Certificate{cer}, RootCAs: caCertPool}

	conn, err := tls.Dial("tcp", nodeRef.TLSAddress, config)
	if err != nil {
		fmt.Printf("TLS Dial to %s failed with error: %v\n", nodeRef.TLSAddress, err)
		return nil, err
	}
	defer conn.Close()

	// read the file in node storage
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Failed to open file: %v\n", err)
		return nil, err
	}
	defer file.Close()

	// write the file into a byte array
	data := make([]byte, 1024)
	_, err = file.Read(data)
	if err != nil {
		fmt.Printf("Failed to read file: %v\n", err)
		return nil, err
	}

	return data, nil
}

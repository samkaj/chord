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

func (node *Node) handleConnection(conn net.Conn) {
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	data := buffer[:n]
	fileName := string(data[:strings.Index(string(data), "\n")])
	data = data[strings.Index(string(data), "\n")+1:]
	receivedFrom := string(data[:strings.Index(string(data), "\n")])
	data = data[strings.Index(string(data), "\n")+1:]

	if fileName[:7] == "backup_" {
		os.Mkdir(node.StoragePath+"/backup-"+receivedFrom, 0777)
		file, err := os.Create(node.StoragePath + "/backup-" + receivedFrom + "/" + fileName[7:])
		if err != nil {
			fmt.Println("Failed to create file: ", err)
		}
		defer file.Close()
		return
	}

	file, err := os.Create(node.StoragePath + "/" + fileName)
	if err != nil {
		fmt.Println("Failed to create file: ", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		fmt.Println("Failed to write to file: ", err)
	}

	if fileName[:7] != "backup_" {
		node.StoreBackups()
	}

}

func (node *Node) StoreBackups() {
	/* files, err := os.ReadDir(node.StoragePath)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Storing backups for node: ", Hash(node.Address))
	log.Println("Node: ", Hash(node.Address), "Files: ", files)

	for _, file := range files {
		fileName := file.Name()
		if fileName[:7] == "backup_" {
			continue
		}
		for _, successor := range node.Successors {
			if successor.Address != node.Address {
				data, err := os.ReadFile(node.StoragePath + "/" + fileName)
				if err != nil {
					log.Fatal(err)
				}
				log.Println("Storing backup: ", fileName, "from: ", node.Address, "to: ", successor.Address)
				backupFileName := "backup_" + fileName
				TLSSend(node, successor, backupFileName, data)
			}
		}
	} */
}

func TLSSend(node *Node, nodeRef NodeRef, fileName string, data []byte) {
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

	header := []byte(fileName + "\n" + node.Address + "\n")
	data = append(header, data...)
	_, err = conn.Write(data)
	if err != nil {
		fmt.Println("TLS Write error: ", err)
	}
}

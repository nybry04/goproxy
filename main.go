package main

import (
	"io"
	"log"
	"net"
	"os"
)

func main() {
	listen := os.Getenv("LISTEN")
	target := os.Getenv("TARGET")

	listener, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %s\n", listen, err)
	}
	defer listener.Close()

	log.Printf("Listening on %s and forwarding to %s\n", listen, target)

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %s\n", err)
			continue
		}
		go handleConnection(clientConn, target)
	}
}

func handleConnection(clientConn net.Conn, targetAddr string) {
	defer clientConn.Close()

	targetConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		log.Printf("Failed to connect to target: %s\n", err)
		return
	}
	defer targetConn.Close()

	go io.Copy(targetConn, clientConn)
	io.Copy(clientConn, targetConn)
}

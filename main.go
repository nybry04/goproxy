package main

import (
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/supabase-community/supabase-go"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type User struct {
	Id       string
	Ip       string
	Username string
	Password string
}

var users map[string]User
var usersMutex sync.Mutex

func fetchUsers(url, token string) {
	client, err := supabase.NewClient(url, token, &supabase.ClientOptions{})
	if err != nil {
		fmt.Println("Cannot initialize client", err)
	}
	data, _, _ := client.From("minecraft").Select("*", "exact", false).Execute()

	localUsers := make(map[string]User)

	jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		id, _ := jsonparser.GetString(value, "id")
		ip, _ := jsonparser.GetString(value, "ip")
		username, _ := jsonparser.GetString(value, "username")
		password, _ := jsonparser.GetString(value, "password")
		localUsers[ip] = User{id, ip, username, password}
	})
	usersMutex.Lock()
	users = localUsers
	usersMutex.Unlock()
}

func main() {
	log.Printf("Version 0.0.8")
	listen := os.Getenv("LISTEN")
	target := os.Getenv("TARGET")
	limbo := os.Getenv("LIMBO")
	supabase_url := os.Getenv("SUPABASE_URL")
	supabase_token := os.Getenv("SUPABASE_TOKEN")

	go func() {
		for {
			fetchUsers(supabase_url, supabase_token)
			time.Sleep(1 * time.Minute)
		}
	}()

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
		log.Printf("New connection from %s\n", clientConn.RemoteAddr().String())
		go handleConnection(clientConn, target, limbo)
	}
}

func handleConnection(clientConn net.Conn, targetAddr string, limboAddr string) {
	defer clientConn.Close()

	ipUser := users[strings.Split(clientConn.RemoteAddr().String(), ":")[0]]

	var target string

	if ipUser.Ip == "" {
		target = limboAddr
	} else {
		target = targetAddr
	}

	targetConn, err := net.Dial("tcp", target)
	if err != nil {
		log.Printf("Failed to connect to target: %s\n", err)
		return
	}
	defer targetConn.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		io.Copy(targetConn, clientConn)
	}()
	io.Copy(clientConn, targetConn)

	<-done
}

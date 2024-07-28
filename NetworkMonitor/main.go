package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os/user"
	"strconv"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/gorilla/websocket"
)

// Define the upgrader which will upgrade the HTTP connection to a WebSocket connection
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins (in production, it's better to set this explicitly)
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Slice to store connected WebSocket clients
var clients []*websocket.Conn
var mu sync.Mutex

// Handler function for WebSocket connection
func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP request to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return
	}
	defer conn.Close()

	mu.Lock()
	clients = append(clients, conn)
	mu.Unlock()

	// Listen for messages from the client
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Client disconnected:", err)
			break
		}
	}

	mu.Lock()
	// Remove the client from the list
	for i, c := range clients {
		if c == conn {
			clients = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	mu.Unlock()
}

// Function to broadcast messages to all connected clients
func broadcastMessage(message string) {
	mu.Lock()
	defer mu.Unlock()

	for _, conn := range clients {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			fmt.Println("Error sending message:", err)
		}
	}
}

var (
	iface   = ""
	snaplen = int32(1600)
	promisc = true
	timeout = pcap.BlockForever
	//filter   = "host 44.228.249.3"
	filter   = "tcp and port 443"
	devFound = false
)

func isRoot() bool {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("[isRoot] Unable to get current user: %s", err)
	}
	return currentUser.Username == "root"
}

func main() {
	// check if running with admin privileges
	if !isRoot() {
		println("Not Running as Admin!")
		return
	}

	// Set up the WebSocket route
	http.HandleFunc("/ws", wsHandler)

	// Start the WebSocket server in a new goroutine
	go func() {
		fmt.Println("WebSocket server starting on :4444")
		if err := http.ListenAndServe("0.0.0.0:4444", nil); err != nil {
			fmt.Println("Error starting server:", err)
		} else {
			fmt.Println("Server running on 4444 port")
		}
	}()

	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Panicln(err)
	}

	println("========== ID =========== Name")
	for id, device := range devices {
		fmt.Printf("========== %d =========== %s\n", id, device.Name)
		if device.Name == iface {
			devFound = true
		}
	}

	var input string
	// Read input from the user
	fmt.Scanln(&input)

	// Parse the input as an integer
	number, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("Error: Please enter a valid integer.")
		return
	}

	for id, device := range devices {
		//fmt.Printf("========== %d =========== %s\n", id, device.Name)
		if id == number {
			devFound = true
			iface = device.Name
		}
	}

	if !devFound {
		log.Panicf("Device named '%s' does not exist\n", iface)
	}

	fmt.Printf("Interface %s selected\n", iface)

	handle, err := pcap.OpenLive(iface, snaplen, promisc, timeout)
	if err != nil {
		log.Panicln(err)
	}
	defer handle.Close()

	//if err := handle.SetBPFFilter(filter); err != nil {
	//	log.Panicln(err)
	//}

	source := gopacket.NewPacketSource(handle, handle.LinkType())
	println("Packet Capture Started")
	for packet := range source.Packets() {
		packetStr := packet.Dump()
		appLayer := packet.ApplicationLayer()
		if appLayer == nil {
			continue
		}
		payload := appLayer.Payload()
		if bytes.Contains(payload, []byte("uname")) || bytes.Contains(payload, []byte("pass")) || bytes.Contains(payload, []byte("USER")) || bytes.Contains(payload, []byte("PASS")) {
			fmt.Print(string(payload))
		}
		broadcastMessage(packetStr)
	}
}

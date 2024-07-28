package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/user"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/gorilla/websocket"
)

type ws_message struct {
	packet_str        string
	src_ip            gopacket.Endpoint
	dst_ip            gopacket.Endpoint
	src_port          gopacket.Endpoint
	dst_port          gopacket.Endpoint
	protocol          gopacket.LayerType
	src_mac           gopacket.Endpoint
	dst_mac           gopacket.Endpoint
	timestamp         time.Time
	app_layer_payload []byte
}

// App struct
type App struct {
	ctx context.Context
}

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
	fmt.Println("Client connected -> ", conn.RemoteAddr())
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

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
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

	a.ctx = ctx

}

func (a *App) IsRoot() bool {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("[isRoot] Unable to get current user: %s", err)
	}
	return currentUser.Username == "root"
}

func (a *App) GetAllDevices() string {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Panicln(err)
	}

	var devicesStr string
	for _, device := range devices {
		devicesStr += fmt.Sprintf("%s,", device.Name)
	}

	return devicesStr
}

func (a *App) StartCapture(iface string, promisc bool, filter string) {
	snaplen := int32(1600 * 2)
	handle, err := pcap.OpenLive(iface, snaplen, promisc, pcap.BlockForever)
	if err != nil {
		log.Panicln(err)
	}
	defer handle.Close()

	if filter != "" {
		if err := handle.SetBPFFilter(filter); err != nil {
			log.Panicln(err)
		}
	}

	source := gopacket.NewPacketSource(handle, handle.LinkType())
	println("Packet Capture Started")
	for packet := range source.Packets() {
		packetStr := packet.Dump()

		// println(packetStr)

		broadcastMessage(packetStr)
	}
}

func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

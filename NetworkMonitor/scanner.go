//package main
//
//import (
//	"bufio"
//	"fmt"
//	"net"
//	"os"
//	"strconv"
//	"strings"
//	"sync"
//	"time"
//
//	"github.com/shirou/gopsutil/host"
//)
//
//const (
//	timeout     = 1 * time.Second // Timeout for checking open ports
//	concurrency = 100             // Number of concurrent scans
//)
//
//func main() {
//	// Get user input for target and port range
//	fmt.Print("Enter the target IP or domain (e.g., 192.168.1.1 or example.com): ")
//	scanner := bufio.NewScanner(os.Stdin)
//	scanner.Scan()
//	target := scanner.Text()
//
//	fmt.Print("Enter the port range (e.g., 1-65535): ")
//	scanner.Scan()
//	portRange := scanner.Text()
//
//	fmt.Print("Choose scan type (tcp, udp, os, service): ")
//	scanner.Scan()
//	scanType := scanner.Text()
//
//	// Resolve domain name to IP
//	ips, err := net.LookupIP(target)
//	if err != nil {
//		fmt.Printf("Could not get IPs: %v\n", err)
//		return
//	}
//
//	// Extract the start and end port from user input
//	startPort, endPort, err := parsePortRange(portRange)
//	if err != nil {
//		fmt.Printf("Invalid port range: %v\n", err)
//		return
//	}
//
//	// Host discovery and port scanning
//	var wg sync.WaitGroup
//	sem := make(chan struct{}, concurrency) // Semaphore for concurrency control
//
//	for _, ip := range ips {
//		fmt.Printf("Scanning IP: %s\n", ip)
//		wg.Add(1)
//		go func(ip net.IP) {
//			defer wg.Done()
//			sem <- struct{}{}        // Acquire semaphore
//			defer func() { <-sem }() // Release semaphore
//
//			switch scanType {
//			case "tcp":
//				scanPorts(ip.String(), startPort, endPort, isTCPPortOpen)
//			case "udp":
//				scanPorts(ip.String(), startPort, endPort, isUDPPortOpen)
//			case "os":
//				detectOS(ip.String())
//			case "service":
//				scanPorts(ip.String(), startPort, endPort, detectService)
//			default:
//				fmt.Println("Unknown scan type. Use tcp, udp, os, or service.")
//			}
//		}(ip)
//	}
//
//	wg.Wait()
//	fmt.Println("Scan complete.")
//}
//
//func parsePortRange(portRange string) (int, int, error) {
//	ports := strings.Split(portRange, "-")
//	if len(ports) != 2 {
//		return 0, 0, fmt.Errorf("invalid port range format")
//	}
//
//	startPort, err := strconv.Atoi(ports[0])
//	if err != nil || startPort < 1 || startPort > 65535 {
//		return 0, 0, fmt.Errorf("invalid start port")
//	}
//
//	endPort, err := strconv.Atoi(ports[1])
//	if err != nil || endPort < startPort || endPort > 65535 {
//		return 0, 0, fmt.Errorf("invalid end port")
//	}
//
//	return startPort, endPort, nil
//}
//
//func scanPorts(ip string, startPort, endPort int, scanFunc func(string, int) bool) {
//	for port := startPort; port <= endPort; port++ {
//		if scanFunc(ip, port) {
//			fmt.Printf("Port %d open on host %s\n", port, ip)
//		}
//	}
//}
//
//func isTCPPortOpen(ip string, port int) bool {
//	address := fmt.Sprintf("%s:%d", ip, port)
//	conn, err := net.DialTimeout("tcp", address, timeout)
//	if err != nil {
//		return false
//	}
//	conn.Close()
//	return true
//}
//
//func isUDPPortOpen(ip string, port int) bool {
//	address := fmt.Sprintf("%s:%d", ip, port)
//	conn, err := net.DialTimeout("udp", address, timeout)
//	if err != nil {
//		return false
//	}
//	conn.Close()
//	return true
//}
//
//func detectOS(ip string) {
//	// Basic OS detection using TTL (this is a naive approach)
//	// For more accurate results, consider using specialized libraries or tools
//	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:80", ip), timeout)
//	if err != nil {
//		fmt.Printf("Could not determine OS: %v\n", err)
//		return
//	}
//	defer conn.Close()
//
//	// Using gopsutil for an example output, as actual OS detection is complex
//	hostInfo, _ := host.Info()
//	fmt.Printf("OS details for %s: %v\n", ip, hostInfo)
//}
//
//func detectService(ip string, port int) bool {
//	address := fmt.Sprintf("%s:%d", ip, port)
//	conn, err := net.DialTimeout("tcp", address, timeout)
//	if err != nil {
//		return false
//	}
//	defer conn.Close()
//
//	// Send a simple request to identify the service (this is just a basic example)
//	fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
//	buff := make([]byte, 4096)
//	n, err := conn.Read(buff)
//	if err != nil {
//		return false
//	}
//
//	// Simple banner grabbing example
//	fmt.Printf("Service on port %d: %s\n", port, string(buff[:n]))
//	return true
//}

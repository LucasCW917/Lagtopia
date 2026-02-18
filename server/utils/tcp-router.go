package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	MaxMessageSize = 1024 * 1024 // 1MB
	HeartbeatInterval = 30 * time.Second
)

type Router struct {
	mu       sync.RWMutex
	services map[string]net.Conn
}

func NewRouter() *Router {
	return &Router{
		services: make(map[string]net.Conn),
	}
}

func main() {
	router := NewRouter()
	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	log.Println("TCP router listening on :9000")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go router.handleConnection(conn)
	}
}

func (r *Router) handleConnection(conn net.Conn) {
	defer conn.Close()

	remote := conn.RemoteAddr().String()
	log.Println("New connection:", remote)

	reader := bufio.NewReader(conn)
	var registeredName string

	for {
		msg, err := readMessage(reader)
		if err != nil {
			if err != io.EOF {
				log.Println("Read error from", remote, ":", err)
			}
			if registeredName != "" {
				r.mu.Lock()
				delete(r.services, registeredName)
				r.mu.Unlock()
				log.Println("Service disconnected:", registeredName)
			}
			return
		}

		response := r.processCommand(conn, msg, &registeredName)
		if response != "" {
			writeMessage(conn, response)
		}
	}
}

func (r *Router) processCommand(conn net.Conn, cmd string, name *string) string {
	cmd = strings.TrimSpace(cmd)
	parts := strings.Fields(cmd)

	if len(parts) == 0 {
		return "ERROR empty command"
	}

	switch strings.ToUpper(parts[0]) {
	case "REGISTER":
		if len(parts) != 2 {
			return "ERROR usage: REGISTER <SERVICE_NAME>"
		}
		service := parts[1]
		r.mu.Lock()
		r.services[service] = conn
		r.mu.Unlock()
		*name = service
		log.Println("Registered service:", service)
		return "OK"

	case "UNREGISTER":
		if len(parts) != 2 {
			return "ERROR usage: UNREGISTER <SERVICE_NAME>"
		}
		service := parts[1]
		r.mu.Lock()
		delete(r.services, service)
		r.mu.Unlock()
		if *name == service {
			*name = ""
		}
		log.Println("Unregistered service:", service)
		return "OK"

	case "SENDTO":
		if len(parts) < 3 {
			return "ERROR usage: SENDTO <SERVICE> \"<payload>\""
		}
		target := parts[1]
		payload := strings.Join(parts[2:], " ")
		payload = strings.Trim(payload, `"`)
		r.mu.RLock()
		targetConn, exists := r.services[target]
		r.mu.RUnlock()
		if !exists {
			return fmt.Sprintf("ERROR service %s not found", target)
		}
		writeMessage(targetConn, payload)
		return "OK"

	case "BROADCAST":
		if len(parts) < 2 {
			return "ERROR usage: BROADCAST \"<payload>\""
		}
		payload := strings.Join(parts[1:], " ")
		payload = strings.Trim(payload, `"`)
		r.mu.RLock()
		for _, c := range r.services {
			writeMessage(c, payload)
		}
		r.mu.RUnlock()
		return "OK"

	case "PING":
		return "PONG"

	case "LIST":
		r.mu.RLock()
		services := []string{}
		for k := range r.services {
			services = append(services, k)
		}
		r.mu.RUnlock()
		return "SERVICES " + strings.Join(services, ",")

	default:
		return "ERROR unknown command"
	}
}

// Read a length-prefixed message
func readMessage(r *bufio.Reader) (string, error) {
	lengthBytes := make([]byte, 4)
	if _, err := io.ReadFull(r, lengthBytes); err != nil {
		return "", err
	}
	length := binary.BigEndian.Uint32(lengthBytes)
	if length > MaxMessageSize {
		return "", fmt.Errorf("message too large")
	}
	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		return "", err
	}
	return string(data), nil
}

// Write a length-prefixed message
func writeMessage(conn net.Conn, msg string) error {
	data := []byte(msg)
	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, uint32(len(data)))
	if _, err := conn.Write(lengthBytes); err != nil {
		return err
	}
	if _, err := conn.Write(data); err != nil {
		return err
	}
	return nil
}

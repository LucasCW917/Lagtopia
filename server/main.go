package main

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"net"
	"os"
)

type Config struct {
	ServerPort int `json:"server_port"`
	TickRate   int `json:"tick_rate"`
	MaxPlayers int `json:"max_players"`
}

func loadConfig() Config {
	file, _ := os.Open("config/configs.json")
	defer file.Close()
	var cfg Config
	json.NewDecoder(file).Decode(&cfg)
	return cfg
}

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

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Server starting. . .")
	log.Println("Loading configuration. . .")
	cfg := loadConfig()
	log.Printf("Configuration loaded: Port=%d, TickRate=%d, MaxPlayers=%d", cfg.ServerPort, cfg.TickRate, cfg.MaxPlayers)

	// Initialize TCP
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	writeMessage(conn, "REGISTER SERVER")
	log.Println("Server registered with TCP router under name 'SERVER'")
}

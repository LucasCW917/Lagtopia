package main

import (
	"encoding/json"
	"log"
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

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Server starting. . .")
	log.Println("Loading configuration. . .")
	cfg := loadConfig()
	log.Printf("Configuration loaded: Port=%d, TickRate=%d, MaxPlayers=%d", cfg.ServerPort, cfg.TickRate, cfg.MaxPlayers)
}

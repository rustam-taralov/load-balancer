package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"net"
	"os"
)

type Config struct {
	ListenAddr string   `yaml:"listen_addr"`
	Servers    []string `yaml:"servers"`
}

var (
	config Config
	count  = 0
)

func main() {

	err := loadConfig("config.yml", &config)
	if err != nil {
		log.Fatalf("Failed to load config: %s", err)
	}

	fmt.Printf("Listening on: %s\n", config.ListenAddr)
	fmt.Printf("Servers: %v\n", config.Servers)

	listener, err := net.Listen("tcp", config.ListenAddr)

	if err != nil {
		log.Fatalf("Failed to listen: %s", err)
	}

	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Fatalf("Failed to close err: %s", err)
		}
	}(listener)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %s", err)
		}

		log.Printf("Accepted connection from: %s", conn.RemoteAddr())
		server := chooseServer()
		go proxy(server, conn)
	}
}

func loadConfig(filename string, cfg *Config) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	decoder := yaml.NewDecoder(file)
	return decoder.Decode(cfg)
}

func chooseServer() string {
	server := config.Servers[count%len(config.Servers)]
	count++
	fmt.Printf("Forwarding to server: %s\n", server)
	return server
}

func proxy(server string, clientConn net.Conn) {
	serverConn, err := net.Dial("tcp", server)
	if err != nil {
		log.Printf("Failed to connect to server %s: %s", server, err)
		return
	}

	defer func(serverConn net.Conn) {
		err := serverConn.Close()
		if err != nil {
			log.Printf("Failed to connect to server %s: %s", server, err)
		}
	}(serverConn)

	defer func(clientConn net.Conn) {
		err := clientConn.Close()
		if err != nil {
			log.Printf("Failed to connect to server %s: %s", server, err)
		}
	}(clientConn)

	routeConnection(serverConn, clientConn)
	routeConnection(clientConn, serverConn)
}

func routeConnection(dst net.Conn, src net.Conn) {
	_, err := io.Copy(dst, src)
	if err != nil {
		log.Printf("Routing error from %s to %s: %s", src.RemoteAddr(), dst.RemoteAddr(), err)
	}
}

package main

import (
	"log"
	"net"
	"strconv"

	"gelf-otlp-forwarder/internal"
)

func main() {

	config, err := internal.LoadConfig()

	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set up TCP listener for GELF messages
	inboundAddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort("", strconv.Itoa(config.InboundPort)))
	if err != nil {
		log.Fatalf("Failed to resolve TCP address: %v", err)
	}

	listener, err := net.ListenTCP("tcp", inboundAddr)
	if err != nil {
		log.Fatalf("Failed to listen on TCP: %v", err)
	}
	defer listener.Close()

	// Set up TCP connection for forwarding
	destAddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(config.OutboundHost, strconv.Itoa(config.OutboundPort)))
	if err != nil {
		log.Fatalf("Failed to resolve destination address: %v", err)
	}
	destConn, err := net.DialTCP("tcp", nil, destAddr)
	if err != nil {
		log.Fatalf("Failed to connect to destination: %v", err)
	}
	defer destConn.Close()

	buf := make([]byte, 65535)
	for {
		// Accept incoming TCP connection
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		defer conn.Close()

		// Read incoming message
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("Error reading TCP: %v", err)
			continue
		}

		// Decompress message
		decompressed, err := internal.Decompress(buf[:n])
		if err != nil {
			log.Printf("Failed to decompress message: %v", err)
			continue
		}

		otlpMessage, err := internal.TransformToOTLP(decompressed)
		if err != nil {
			log.Printf("Failed to transform GELF to OTLP: %v", err)
			continue
		}

		// Forward message
		if err := internal.ForwardMessage(otlpMessage, destConn); err != nil {
			log.Printf("Failed to forward message: %v", err)
		}

	}
}

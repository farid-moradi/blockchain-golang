package main

import (
	"flag"
	"log"
)

func init() {
	log.SetPrefix("Blockchain: ")
}

func main() {
	port := flag.Uint("port", 5000, "TCP Port for Blockchain Server")
	flag.Parse()
	server := NewBlockchainServer(uint16(*port))
	server.Run()
}

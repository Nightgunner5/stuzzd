package main

import (
	"flag"
	"github.com/Nightgunner5/stuzzd/networking"
	"log"
	"net"
	"time"
)

var flagHostPort = flag.String("hostport", ":25565", "The host and port to listen on. Blank host means listening on all interfaces.")

const TICK = time.Second / 20

func main() {
	flag.Parse()

	ln, err := net.Listen("tcp", *flagHostPort)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Now listening on ", *flagHostPort)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Print("Error while accepting a connection: ", err)
				continue
			}
			networking.RegisterEntity(networking.HandlePlayer(conn))
		}
	}()

	for {
		time.Sleep(TICK)
	}
}

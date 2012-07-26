package main

import (
	"flag"
	"github.com/Nightgunner5/stuzzd/config"
	"github.com/Nightgunner5/stuzzd/networking"
	"github.com/Nightgunner5/stuzzd/protocol"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"
)

var flagHostPort = flag.String("hostport", ":25565", "The host and port to listen on. Blank host means listening on all interfaces.")
var flagCPUProfile = flag.String("cpuprofile", "", "write cpu profile to file")
var flagMemProfile = flag.String("memprofile", "", "write memory profile to file")

var flagNumSlots = flag.Int("numslots", 10, "The maximum number of players allowed on this server at a time.")

const TICK = time.Second / 20

func main() {
	flag.Parse()

	if *flagCPUProfile != "" {
		log.Print("Profiling to file ", *flagCPUProfile, " started.")
		f, err := os.Create(*flagCPUProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	go networking.InitSpawnArea()

	ln, err := net.Listen("tcp", *flagHostPort)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Now listening on ", *flagHostPort)

	config.NumSlots = uint8(*flagNumSlots)

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

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT)
		<-ch
		log.Print("Saving ALL the data!")
		networking.SaveAllTheThings()
		if *flagCPUProfile != "" {
			log.Print("Finishing up profile information...")
			pprof.StopCPUProfile()
		}
		if *flagMemProfile != "" {
			log.Print("Writing memory profile...")
			f, err := os.Create(*flagMemProfile)
			if err != nil {
				log.Fatal(err)
			}
			pprof.WriteHeapProfile(f)

		}
		os.Exit(0)
	}()

	for {
		time.Sleep(TICK)
		config.Tick++
		if config.Tick%100 == 0 {
			networking.SendToAll(protocol.TimeUpdate{Time: config.Tick})
		}
	}
}

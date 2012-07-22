package main

import (
	"flag"
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
		go func() {
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, syscall.SIGINT)
			<-ch
			log.Print("Finishing up profile information...")
			pprof.StopCPUProfile()
			os.Exit(0)
		}()
	}

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

	var curTick uint64
	for {
		time.Sleep(TICK)
		curTick++
		if curTick%100 == 0 {
			networking.SendToAll(protocol.TimeUpdate{Time: curTick})
		}
	}
}

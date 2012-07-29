package config

import (
	"encoding/json"
	"log"
	"os"
)

var Tick uint64

type Configuration struct {
	NumSlots          uint64
	ServerDescription string
}

var Config Configuration

func (c *Configuration) Save() {
	f, err := os.Create("stuzzd.conf")
	if err != nil {
		log.Print(err)
	}
	defer f.Close()

	data, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		log.Print(err)
	}

	_, err = f.Write(data)
	if err != nil {
		log.Print(err)
	}
}

func init() {
	// Defaults
	Config.NumSlots = 10
	Config.ServerDescription = "StuzzHosting is Best Hosting"

	// Read the file
	f, err := os.Open("stuzzd.conf")
	if err != nil {
		log.Print(err)
	}
	if f == nil {
		Config.Save()
		return
	}
	defer f.Close()

	r := json.NewDecoder(f)
	err = r.Decode(&Config)
	if err != nil {
		// If the config file has errors, don't continue with possibly unwanted operation.
		log.Fatal(err)
	}
}

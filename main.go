package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"

	"github.com/ldkingvivi/udp_proxy/udp"
)

var (
	configFile = flag.String("config", "", "Configuration file to use")
)

func main() {

	flag.Parse()

	if *configFile == "" {
		fmt.Fprintln(os.Stderr, "Missing configuration file")
		flag.PrintDefaults()
		os.Exit(1)
	}

	content, _ := ioutil.ReadFile(*configFile)

	var conf udp.UDPConfig
	err := json.Unmarshal(content, &conf)
	if err != nil {
		fmt.Print("Error:", err)
	}

	fmt.Println(conf)

	u, err := udp.InitUDPServer(conf)
	if err != nil {
		log.Fatal(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	go func() {
		<-sigChan
		u.Stop()
		fmt.Println("Exiting ... ")
		os.Exit(0)
	}()

	log.Println("starting our awesome Shield Collectd Proxy...")
	u.Start()
	//time.Sleep(1800 * time.Second)

}

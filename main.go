package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func init() {
	log.SetPrefix("Tcp Server >")
}

var server *TcpServer

func runLoop() {
	sigs := make(chan os.Signal, 1)
	defer close(sigs)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	s := <-sigs
	log.Println("Termination requested:", s.String())
	server.Stop()
}

func main() {
	var header string = `
___________                    _________                                      
\__    ___/____  ______       /   _____/  ____ _______ ___  __  ____ _______  
  |    | _/ ___\ \____ \      \_____  \ _/ __ \\_  __ \\  \/ /_/ __ \\_  __ \ 
  |    | \  \___ |  |_> >     /        \\  ___/ |  | \/ \   / \  ___/ |  | \/ 
  |____|  \___  >|   __/     /_______  / \___  >|__|     \_/   \___  >|__|    
              \/ |__|                \/      \/                    \/         
                                                                              
`
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println(header)
	fmt.Println(strings.Repeat("-", 80))
	port := flag.Int("port", 0, "Server port")

	flag.Parse()

	server = NewTcpServer(fmt.Sprintf(":%d", *port))
	log.Printf("Tcp Server %s is running. <Ctrl-C to stop>", server.GetAddress())
	runLoop()
}

package main

import (
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type TcpServer struct {
	listener net.Listener
	quit     chan interface{}
	wg       sync.WaitGroup
}

func (srv *TcpServer) GetAddress() string {
	return srv.listener.Addr().String()
}

func NewTcpServer(ip_addr_port string) *TcpServer {
	tcpSrv := &TcpServer{
		quit: make(chan interface{}),
	}
	var err error
	tcpSrv.listener, err = net.Listen("tcp", ip_addr_port)
	if err != nil {
		log.Printf("Tcp Server %s failed: %+v", ip_addr_port, err)
		log.Fatal(err)
	}
	// add wg for the go routine 'run'
	tcpSrv.wg.Add(1)
	go tcpSrv.run()
	return tcpSrv
}

func (srv *TcpServer) run() {
	defer srv.wg.Done() // indicate to the master that the routine has terminated
	for {
		conn, err := srv.listener.Accept()
		if err != nil {
			select {
			case <-srv.quit: // accept fails on intended closed listener
				return
			default:
				log.Printf("Tcp Server accept failed: %+v", err)
			}
		} else {
			srv.wg.Add(1)
			go func() {
				srv.handleConection(conn)
				srv.wg.Done()
			}()
		}
	}
}

func (srv *TcpServer) Stop() {
	log.Println("Tcp Server is stopping...")
	close(srv.quit)      // pre-awake the select accept logic
	srv.listener.Close() // cause accept to fail
	srv.wg.Wait()        //waiting for all the routines to stop
	log.Println("Tcp Server has stopped.")
}

func (srv *TcpServer) handleConection(conn net.Conn) {
	defer conn.Close()        // close the client connection on exit
	buf := make([]byte, 4096) // tmp buf 4KB - conf
	var offset int = 0
	for {
		nb, err := conn.Read(buf[offset:])
		if err != nil && err != io.EOF {
			log.Printf("Tcp Server client connection handler failed: %+v", err)
			return
		}
		if nb == 0 {
			break
		}
		offset += nb
		log.Printf("Tcp Server received from %v: %s", conn.RemoteAddr(), string(buf[offset:]))
	}
	log.Printf("Tcp Server received from %v: %s", conn.RemoteAddr(), string(buf[:]))
}

func (s *TcpServer) handleConectionGracious(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 4096) // inner buffer size - conf
	var offset int = 0
	flag := true
	for flag {
		select {
		case <-s.quit:
			return
		default:
			conn.SetDeadline(time.Now().Add(200 * time.Millisecond)) // timeout - conf
			nb, err := conn.Read(buf[offset:])
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					continue
				} else if err != io.EOF {
					log.Printf("Tcp Server handler read error: %+v", err)
					return
				}
			}
			if nb == 0 {
				flag = false
				break
			}
			offset += nb
			log.Printf("Tcp Server received from %v: %s", conn.RemoteAddr(), string(buf[offset:]))
		}
	}
	log.Printf("Tcp Server received from %v: %s", conn.RemoteAddr(), string(buf[:]))
}

package main

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

type Server struct {
	lisenter net.Listener
	msgChan  chan net.Conn
	quitChan chan bool
	wg       sync.WaitGroup
}

func (s *Server) acceptConnections() {
	defer s.wg.Done()
	for {
		select {
		case <-s.quitChan:
			return
		default:
			conn, err := s.lisenter.Accept()
			if err != nil {
				fmt.Printf("failed to establish a TCP connection: %s", err.Error())
				continue
			}

			s.msgChan <- conn
		}
	}
}

func (s *Server) handleConnection() {
	defer s.wg.Done()

	for {
		select {
		case con := <-s.msgChan:
			fmt.Printf("recieved a message\n")
			buffer := make([]byte, 1024)
			con.Read(buffer)
			fmt.Printf("the message is: %s", buffer)
			con.Close()
		case <-s.quitChan:
			return
		}
	}
}

func (s *Server) startUp() {
	s.wg.Add(2)
	go s.acceptConnections()
	go s.handleConnection()
}

func (s *Server) shutDown() {
	s.quitChan <- true

	done := make(chan bool)
	go func() {
		s.wg.Done()
		done <- true
	}()

	select {
	case <-done:
		s.lisenter.Close()
		fmt.Printf("we are done here\n")
		return
	case <-time.After(time.Second):
		panic("Something is really wrong")
	}
}

func connectClientToTcp() {
	time.Sleep(time.Millisecond * 500)

	for i := 0; i < 100; i++ {

		connection, err := net.Dial("tcp", ":8080")
		if err != nil {
			fmt.Printf("failed to create a TCP connection\n")
		}
		msg := []byte("Hello From the other side" + strconv.FormatInt(int64(i), 10))

		connection.Write(msg)
	}

	time.Sleep(time.Second * 2)
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Printf("Failed to start TCP server: %s", err.Error())
	}

	server := &Server{lisenter: listener, msgChan: make(chan net.Conn), quitChan: make(chan bool)}
	server.startUp()

	go connectClientToTcp()

	time.Sleep(time.Second * 6)
	server.shutDown()
}

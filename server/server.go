package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"simplechat/protocol"
	"sync"
)

type server struct {
	onlineClients map[int]*Client
	mtx           sync.RWMutex
	nextIdCanUsed int
}

func (s *server) Init() {
	s.nextIdCanUsed = 0
	s.onlineClients = make(map[int]*Client)
}

// 处理写
func (s *server) HandleWrite(c *Client) {
	for {
		// 交换缓冲区
		c.wbuffMtx.Lock()
		tmpBuff := c.writeBuffer
		c.writeBuffer = nil
		c.wbuffMtx.Unlock()
		// 写缓冲区
		n := len(tmpBuff)
		hasWriten := 0
		for hasWriten < n {
			writen, err := (*(c.conn)).Write(tmpBuff[hasWriten:])
			if err != nil {
				fmt.Println("HandleWrite: ", (*(c.conn)).RemoteAddr().String(), err)
				c.mtx.Lock()
				c.online = false
				c.mtx.Unlock()
			} else {
				hasWriten += writen
			}
		}
		c.mtx.RLock()
		if !c.online {
			c.mtx.RUnlock()
			break
		}
		c.mtx.RUnlock()
	}
	(*(c.conn)).Close()
}

func (s *server) HandleRead(c *Client) {
	for {
		tmpBuffer := make([]byte, 1024)
		n, err := (*(c.conn)).Read(tmpBuffer)
		if err != nil {
			c.mtx.Lock()
			c.online = false
			c.mtx.Unlock()
			s.mtx.Lock()
			delete(s.onlineClients, c.id)
			s.mtx.Unlock()
			break
		} else {
			c.readBuffer = append(c.readBuffer, tmpBuffer[:n]...)
		}
		readBuffer, isMsg, senderId, msgType, msg := protocol.Depack(c.readBuffer)
		c.readBuffer = readBuffer
		if isMsg {
			switch msgType {
			case protocol.NAME:
				if senderId == c.id {
					c.name = msg
				}
			case protocol.MSG:
				if senderId == c.id {
					chatMsg := protocol.Enpack(senderId, msgType, []byte(c.name+": "+msg))
					s.mtx.RLock()
					for id, otherc := range s.onlineClients {
						if id != c.id {
							otherc.wbuffMtx.Lock()
							otherc.writeBuffer = append(otherc.writeBuffer, chatMsg...)
							otherc.wbuffMtx.Unlock()
						}
					}
					s.mtx.RUnlock()
				}
			}
		}
	}
}

func (s *server) RunServer(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("Listen: ", err)
		os.Exit(1)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		} else {
			c := new(Client)
			c.conn = &conn
			c.id = s.nextIdCanUsed
			c.online = true
			s.nextIdCanUsed++
			s.onlineClients[c.id] = c
			fmt.Println("Connection:", conn.RemoteAddr().String(), "id:", c.id)
			// 交给goroutine处理读写
			idMsg := protocol.Enpack(0, protocol.ID, protocol.IntToBytes(c.id))
			c.writeBuffer = append(c.writeBuffer, idMsg...)
			go s.HandleRead(c)
			go s.HandleWrite(c)
		}
	}
}

type Client struct {
	id          int
	name        string
	readBuffer  []byte
	writeBuffer []byte
	wbuffMtx    sync.Mutex
	conn        *net.Conn
	online      bool
	mtx         sync.RWMutex
}

func main() {
	addr := "localhost:6666"
	s := new(server)
	s.Init()
	s.RunServer(addr)
}

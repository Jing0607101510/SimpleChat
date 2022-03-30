package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"simplechat/protocol"
	"strings"
)

type Client struct {
	id           int
	name         string
	readBuffer   []byte
	writeChannel chan string
	conn         *net.TCPConn
}

func GetUserName() string {
	fmt.Println("请输入您的用户名：")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	userName := strings.Trim(scanner.Text(), " ")
	return userName
}

func Writer(client *Client) {
	for {
		msg := <-client.writeChannel
		buffer := protocol.Enpack(client.id, protocol.MSG, []byte(msg))
		hasWriten := 0
		for hasWriten < len(buffer) {
			n, err := client.conn.Write(buffer[hasWriten:])
			if err != nil {
				fmt.Println("Writer: ", err)
				os.Exit(1)
			}
			hasWriten += n
		}
	}
}

func Reader(client *Client) {
	for {
		buf := make([]byte, 1024)
		n, err := client.conn.Read(buf)
		if err != nil {
			fmt.Println("Reader: ", err)
			os.Exit(1)
		}
		client.readBuffer = append(client.readBuffer, buf[:n]...)
		// 处理逻辑
		readBuffer, isMsg, _, MsgType, msg := protocol.Depack(client.readBuffer)
		client.readBuffer = readBuffer
		if isMsg {
			switch MsgType {
			case protocol.ID:
				client.id = protocol.BytesToInt([]byte(msg))
			case protocol.MSG:
				fmt.Println(msg)
				fmt.Print(">> ")
			}
		}
	}
}

func ReceiveUserID(client *Client) {
	needReceive := protocol.SenderID + protocol.MsgType + protocol.MsgLen + protocol.SenderID
	hasReceived := 0

	tmpBuffer := make([]byte, 1024)
	for hasReceived < needReceive {
		n, err := client.conn.Read(tmpBuffer)
		if err != nil {
			fmt.Println("ReceiveUserID: ", err)
			os.Exit(1)
		}
		client.readBuffer = append(client.readBuffer, tmpBuffer[:n]...)
		hasReceived += n
	}
	readBuffer, ok, _, _, id := protocol.Depack(client.readBuffer)
	client.readBuffer = readBuffer
	if ok {
		client.id = protocol.BytesToInt([]byte(id))
	} else {
		fmt.Println("ReceiveUserID: 获取id失败")
	}
}

func SendUserName(client *Client) {
	nameBytes := []byte(client.name)
	buffer := protocol.Enpack(client.id, protocol.NAME, nameBytes)
	hasWriten := 0
	for hasWriten < len(nameBytes) {
		n, err := client.conn.Write(buffer[hasWriten:])
		if err != nil {
			fmt.Println("SendUserName: ", err)
			os.Exit(1)
		}
		hasWriten += n
	}
}

func main() {
	client := new(Client)
	userName := GetUserName()
	client.name = userName
	client.writeChannel = make(chan string, 16)

	server := "localhost:6666"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", server)
	if err != nil {
		fmt.Println("ResolveTCPAddr: ", err)
		os.Exit(1)
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println("DialTCP: ", err)
		os.Exit(1)
	}
	defer conn.Close()
	client.conn = conn

	ReceiveUserID(client)
	SendUserName(client)

	go Reader(client)
	go Writer(client)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(">> ")
		scanner.Scan()
		msg := scanner.Text()
		client.writeChannel <- msg
	}
}

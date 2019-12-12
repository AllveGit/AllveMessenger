package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"sync"
)

const (
	MAX_USER_NUM = 64
)

var MessageChannel chan string
var QuitChannel chan Client

type messagePacket struct {
	messageBufferSize [8]byte
	message           string
}

type Client struct {
	name    string
	connect net.Conn
	end     bool
}

var serverAndPort string = "127.0.0.1:9999"
var users map[Client]bool
var curUserNum int = 0

var CriticalSection sync.WaitGroup

func ReadMessage(user *Client) string {

	var messageBufferSize [8]byte
	var packetSize int

	var messageData string = ""

	_, bufErr := user.connect.Read([]byte(messageBufferSize[:]))

	if !errorCheck(bufErr) {
		user.end = true
	}

	packetSize = int(binary.BigEndian.Uint64(messageBufferSize[:]))

	recvBuf := make([]byte, packetSize)

	for 0 < packetSize {
		_, msgErr := user.connect.Read(recvBuf)

		if !errorCheck(msgErr) {
			user.end = true
			break
		}

		messageData = string(recvBuf)

		packetSize -= len(messageData)
	}

	return messageData
}

func SendMessage(user *Client, message string) {
	var packet messagePacket

	var messageBufferSize [8]byte
	var packetSize int

	packet.message = message
	packetSize = len(packet.message)
	fmt.Println(packetSize)

	binary.BigEndian.PutUint64(messageBufferSize[:], uint64(packetSize))
	_, bufErr := user.connect.Write(messageBufferSize[:])

	if !errorCheck(bufErr) {
		user.end = true
		return
	}

	_, msgErr := user.connect.Write([]byte(packet.message))

	if !errorCheck(msgErr) {
		user.end = true
		return
	}
}

func BroadCast(message string) {
	for key, _ := range users {
		SendMessage(&key, message)
	}
}

func QuitClientCheck() {
	for {
		for QuitClient := range QuitChannel {
			delete(users, QuitClient)
			QuitClient.end = true
		}

	}
}

func MessageCheck() {
	for {
		for Messages := range MessageChannel {
			BroadCast(Messages)
		}
	}
}

func ClientReadMessage(user *Client) {
	defer user.connect.Close()

	for !user.end {
		readedMessage := ReadMessage(user)

		if strings.Compare(readedMessage, "quit\r\n") == 0 {
			SendMessage(user, "quit")
			QuitChannel <- *user
		} else if strings.HasPrefix(readedMessage, "[N]") {
			readedMessage = strings.TrimLeft(readedMessage, "[N]")
			user.name = readedMessage
		} else {
			readedMessage = user.name + " : " + readedMessage
			fmt.Println(readedMessage)
			MessageChannel <- readedMessage

			// BroadCast(readedMessage)
		}
	}
}

func errorCheck(errorLog error) bool {

	if errorLog != nil {
		fmt.Println(errorLog)
		return false
	}

	return true
}

func StartServer(_serverAndPort string) net.Listener {

	users = make(map[Client]bool)

	MessageChannel = make(chan string, MAX_USER_NUM)
	QuitChannel = make(chan Client, MAX_USER_NUM)

	server, err := net.Listen("tcp", _serverAndPort)

	if !errorCheck(err) {
		return nil
	}

	return server
}

func main() {
	server := StartServer(serverAndPort)

	defer server.Close()

	go QuitClientCheck()
	go MessageCheck()

	for {
		newClient, err := server.Accept()

		if !errorCheck(err) {
			return
		}

		fmt.Println("새로운 클라이언트가 서버에 들어왔습니다")

		processClient := Client{"", newClient, false}

		users[processClient] = true

		go ClientReadMessage(&processClient)
	}
}

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

const (
	serverAndPort = "127.0.0.1:9999"

	STATE_LOBBY  = 1
	STATE_ROOMIN = 2
)

type MessagePacket struct {
	HeaderSize int
	Message    string
}

type Client struct {
	connect net.Conn
	state   int
	quit    bool
}

func ErrorCheck(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func quit(client *Client) {
	time.Sleep(time.Second * 3)
	client.quit = true
}

func SendPacket(client *Client) {

	var message string = ""

	for !client.quit {
		in := bufio.NewReader(os.Stdin)
		line, inputErr := in.ReadString('\n')

		ErrorCheck(inputErr)

		message = line

		if strings.Compare(message, "quit\r\n") == 0 {
			go quit(client)
		} else if strings.Compare(message, "1\r\n") == 0 {
			fmt.Println("생성할 채팅방 이름을 입력해주세요 : ")

			roomIn := bufio.NewReader(os.Stdin)
			roomName, inputErr := roomIn.ReadString('\n')
			ErrorCheck(inputErr)

			message = "[ROOMCREATE]" + roomName
		} else if strings.Compare(message, "2\r\n") == 0 {
			fmt.Println("채팅방 리스트들을 불러옵니다")

			message = "[ROOMFIND]"
		} else if strings.Compare(message, "3\r\n") == 0 {
			fmt.Println("들어갈 채팅방 이름을 입력해주세요 : ")

			roomIn := bufio.NewReader(os.Stdin)
			roomName, inputErr := roomIn.ReadString('\n')
			ErrorCheck(inputErr)

			message = "[ROOMIN]" + roomName
		}

		messagePacket := MessagePacket{len(message), message}

		jsonBytes, encodingErr := json.Marshal(messagePacket)

		ErrorCheck(encodingErr)

		_, msgErr := client.connect.Write(jsonBytes)

		ErrorCheck(msgErr)
	}

}

func ReadPacket(client *Client) {

	var headerSize int = 0
	var messageData string = ""

	readData := make([]byte, 4096)
	messagePacket := MessagePacket{0, ""}

	for !client.quit {
		readSize, msgErr := client.connect.Read(readData)

		ErrorCheck(msgErr)

		decodingErr := json.Unmarshal(readData[:readSize], &messagePacket)

		headerSize = messagePacket.HeaderSize

		ErrorCheck(decodingErr)

		messageData += messagePacket.Message

		headerSize -= readSize

		for headerSize > 0 {
			excReadSize, excMsgErr := client.connect.Read(readData)

			ErrorCheck(excMsgErr)

			decodingErr := json.Unmarshal(readData[:excReadSize], &messagePacket)

			ErrorCheck(decodingErr)

			messageData += messagePacket.Message

			headerSize -= excReadSize
		}

		fmt.Println(messageData)

		messageData = ""
	}
}

func TryConnect() Client {
	connect, err := net.Dial("tcp", serverAndPort)

	ErrorCheck(err)

	connectClient := Client{connect, 1, false}

	return connectClient
}

func main() {

	client := TryConnect()

	defer client.connect.Close()

	fmt.Println("서버에 연결되었습니다")

	fmt.Println("1. 채팅방 생성  2. 채팅방 탐색  3. 채팅방 들어가기")

	go SendPacket(&client)

	go ReadPacket(&client)

	for !client.quit {

	}
}

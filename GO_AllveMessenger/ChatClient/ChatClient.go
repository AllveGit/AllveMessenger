package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strings"
)

var nickname string
var serverAndPort string = "127.0.0.1:9999"
var quit bool = false

type messagePacket struct {
	messageBufferSize [8]byte
	message           string
}

func errorCheck(errorLog error) bool {

	if errorLog != nil {
		fmt.Println(errorLog)
		return false
	}

	return true
}

func SendMessage(connect net.Conn, sendChannelDone *chan bool) {

	var packet messagePacket

	var messageBufferSize [8]byte
	var packetSize int

	var bNickName bool = true

	for !quit {
		// fmt.Scan(&packet.message)
		in := bufio.NewReader(os.Stdin)
		line, err := in.ReadString('\n')

		if !errorCheck(err) {
			quit = true
			break
		}

		if bNickName {
			packet.message = nickname
			bNickName = false
		} else {
			packet.message = line
		}

		packetSize = len(packet.message)
		// fmt.Println(packetSize)

		binary.BigEndian.PutUint64(messageBufferSize[:], uint64(packetSize))
		_, bufErr := connect.Write(messageBufferSize[:])

		if !errorCheck(bufErr) {
			quit = true
			break
		}

		_, msgErr := connect.Write([]byte(packet.message))

		if !errorCheck(msgErr) {
			quit = true
			break
		}

		if strings.Compare(packet.message, "quit\r\n") == 0 {
			quit = true
			break
		}
	}

	*sendChannelDone <- true
}

func ReadMessage(connect net.Conn, readChannelDone *chan bool) {

	var messageBufferSize [8]byte
	var packetSize int

	for !quit {
		_, bufErr := connect.Read(messageBufferSize[:])

		if !errorCheck(bufErr) {
			quit = true
			break
		}

		packetSize = int(binary.BigEndian.Uint64(messageBufferSize[:]))
		var messageData string = ""

		recvBuf := make([]byte, packetSize)

		for 0 < packetSize {
			_, msgErr := connect.Read(recvBuf)

			if !errorCheck(msgErr) {
				quit = true
				break
			}

			messageData = string(recvBuf)

			packetSize -= len(recvBuf)
		}

		fmt.Println(messageData)
	}

	*readChannelDone <- true
}

func TryConnect(_serverAndPort string) net.Conn {
	connect, err := net.Dial("tcp", _serverAndPort)

	if !errorCheck(err) {
		return nil
	}

	return connect
}

func main() {
	fmt.Print("닉네임을 입력하세요 : ")
	fmt.Scanln(&nickname)
	nickname = "[N]" + nickname

	connect := TryConnect(serverAndPort)

	defer connect.Close()

	fmt.Println("서버에 연결 성공하였습니다")

	sendChannelEnd := make(chan bool)
	readChannelEnd := make(chan bool)

	go SendMessage(connect, &sendChannelEnd)
	go ReadMessage(connect, &readChannelEnd)

	<-sendChannelEnd
	<-readChannelEnd

	close(sendChannelEnd)
	close(readChannelEnd)

	return
}

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	MAX_USER_NUM  = 64
	serverAndPort = "127.0.0.1:9999"
)

type Room struct {
	roomID     int
	roomName   string
	roomUsers  map[int]Client
	msgChannel chan string
}

type Client struct {
	netID   int
	inRoom  bool
	roomID  int
	curRoom *Room
	connect net.Conn
}

type Server struct {
	listener                net.Listener
	users                   map[int]Client
	rooms                   map[string]Room
	clientQuitChannel       chan Client
	roomCreateChannel       chan string
	allRoomBroadcastChannel chan string
}

type MessagePacket struct {
	HeaderSize int
	Message    string
}

func ErrorCheck(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func ReadPacket(user *Client) string {
	var headerSize int = 0
	var messageData string = ""

	readData := make([]byte, 100)
	messagePacket := MessagePacket{0, ""}

	readSize, msgErr := user.connect.Read(readData)

	ErrorCheck(msgErr)

	decodingErr := json.Unmarshal(readData[:readSize], &messagePacket)

	ErrorCheck(decodingErr)

	headerSize = messagePacket.HeaderSize

	messageData += messagePacket.Message

	headerSize -= readSize

	for headerSize > 0 {
		excReadSize, excMsgErr := user.connect.Read(readData)

		ErrorCheck(excMsgErr)

		decodingErr := json.Unmarshal(readData[:excReadSize], &messagePacket)

		ErrorCheck(decodingErr)

		messageData += messagePacket.Message

		headerSize -= excReadSize
	}

	fmt.Println(messageData)

	return messageData

}

func SendPacket(user *Client, message string) {

	messagePacket := MessagePacket{len(message), message}

	jsonBytes, encodingErr := json.Marshal(messagePacket)

	ErrorCheck(encodingErr)

	_, msgErr := user.connect.Write(jsonBytes)

	ErrorCheck(msgErr)
}

func ServerChannelHandle(server *Server) {
	for {
		select {
		case quitClient := <-server.clientQuitChannel:
			delete(server.users, quitClient.netID)

			if quitClient.inRoom == true {
				delete(quitClient.curRoom.roomUsers, quitClient.roomID)
			}

			quitClient.connect.Close()

			for quitClient = range server.clientQuitChannel {
				delete(server.users, quitClient.netID)

				if quitClient.inRoom == true {
					delete(quitClient.curRoom.roomUsers, quitClient.roomID)
				}

				quitClient.connect.Close()
			}
		// case broadCastMessage := <-server.allRoomBroadcastChannel:
		case createRoomName := <-server.roomCreateChannel:
			roomUsers := make(map[int]Client)
			roomMsgChannel := make(chan string, MAX_USER_NUM)
			newRoom := Room{0, createRoomName, roomUsers, roomMsgChannel}
			server.rooms[createRoomName] = newRoom
		}
	}
}

func UpdateUsers(server *Server) {
	for {
		for _, client := range server.users {
			UpdateUser(&client, server)
		}
	}
}

func UpdateUser(user *Client, server *Server) {
	messagePacket := ReadPacket(user)

	if strings.HasPrefix(messagePacket, "[ROOMCREATE]") {
		messagePacket = strings.TrimLeft(messagePacket, "[ROOMCREATE]")
		server.roomCreateChannel <- messagePacket
	} else if strings.HasPrefix(messagePacket, "[ROOMFIND]") {
		var roomListInfo string = ""
		var roomIndex int = 0

		for roomName, _ := range server.rooms {
			var roomInfo string = ""
			roomInfo = fmt.Sprintf("%d - %s\n", roomIndex, roomName)
			roomListInfo += roomInfo
			roomIndex += 1
		}

		SendPacket(user, roomListInfo)

		roomListInfo = ""
		roomIndex = 0
	}
}

func ServerStart() Server {
	listener, err := net.Listen("tcp", serverAndPort)

	ErrorCheck(err)

	users := make(map[int]Client)
	rooms := make(map[string]Room)
	quitChannel := make(chan Client, MAX_USER_NUM)
	roomCreateChannel := make(chan string)
	allRoomBroadcastChannel := make(chan string, MAX_USER_NUM)

	server := Server{listener, users, rooms, quitChannel, roomCreateChannel, allRoomBroadcastChannel}

	return server
}

func main() {
	var iUserNum int = 0

	server := ServerStart()

	go ServerChannelHandle(&server)
	go UpdateUsers(&server)

	for {
		tryConnectClient, err := server.listener.Accept()

		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("새로운 클라이언트가 접속하였습니다")

		newClient := Client{iUserNum, false, 0, nil, tryConnectClient}

		server.users[newClient.netID] = newClient

		iUserNum += 1
	}
}

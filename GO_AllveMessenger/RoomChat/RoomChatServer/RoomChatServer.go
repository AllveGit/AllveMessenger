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
	RoomName   string
	RoomUsers  map[int]Client
	MsgChannel chan string
}

type Client struct {
	netID   int
	inRoom  bool
	end     bool
	curRoom Room
	connect net.Conn
}

type Server struct {
	listener                net.Listener
	Users                   map[int]Client
	Rooms                   map[string]Room
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

	readData := make([]byte, 4096)
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

func BroadCast(message string, room *Room) {
	for _, user := range room.RoomUsers {
		SendPacket(&user, message)
	}
}

func RoomsChannelHandle(room *Room, server *Server) {
	for {
		for message := range room.MsgChannel {
			BroadCast(message, room)
		}
	}
}

func ServerChannelHandle(server *Server) {
	for {
		select {
		case quitClient := <-server.clientQuitChannel:

			SendPacket(&quitClient, "quit")

			delete(server.Users, quitClient.netID)

			if quitClient.inRoom == true {
				delete(quitClient.curRoom.RoomUsers, quitClient.netID)
			}

			quitClient.connect.Close()

			for quitClient = range server.clientQuitChannel {
				SendPacket(&quitClient, "quit")

				delete(server.Users, quitClient.netID)

				if quitClient.inRoom == true {
					delete(quitClient.curRoom.RoomUsers, quitClient.netID)
				}

				quitClient.connect.Close()
			}
		// case broadCastMessage := <-server.allRoomBroadcastChannel:
		case createRoomName := <-server.roomCreateChannel:
			roomUsers := make(map[int]Client)
			roomMsgChannel := make(chan string, MAX_USER_NUM)
			newRoom := Room{createRoomName, roomUsers, roomMsgChannel}
			server.Rooms[createRoomName] = newRoom

			go RoomsChannelHandle(&newRoom, server)
		}
	}
}

func UpdateUser(user *Client, server *Server) {
	for !user.end {
		messagePacket := ReadPacket(user)

		if strings.HasPrefix(messagePacket, "quit") {
			user.end = true
			server.clientQuitChannel <- *user
		} else if strings.HasPrefix(messagePacket, "[ROOMCREATE]") { //방 생성 명령
			messagePacket = strings.TrimLeft(messagePacket, "[ROOMCREATE]")
			server.roomCreateChannel <- messagePacket
		} else if strings.HasPrefix(messagePacket, "[ROOMFIND]") { //방 리스트 불러오기 명령
			var roomListInfo string = ""
			var roomIndex int = 0

			for roomName, _ := range server.Rooms {
				var roomInfo string = ""
				roomInfo = fmt.Sprintf("%d - %s\n", roomIndex, roomName)
				roomListInfo += roomInfo
				roomIndex += 1
			}

			SendPacket(user, roomListInfo)

			roomListInfo = ""
			roomIndex = 0
		} else if strings.HasPrefix(messagePacket, "[ROOMIN]") { //방 들어가기 명령
			messagePacket = strings.TrimLeft(messagePacket, "[ROOMIN]")

			_, isContain := (server.Rooms[messagePacket])

			if !isContain {
				SendPacket(user, "그런 이름을 가진 채팅방이 없습니다")
				return
			}

			if user.inRoom == true {
				delete(user.curRoom.RoomUsers, user.netID)
			}

			user.curRoom = (server.Rooms[messagePacket])
			user.inRoom = true
			server.Rooms[messagePacket].RoomUsers[user.netID] = *user

			var roomInNotice string = server.Rooms[messagePacket].RoomName + " 채팅방에 입장하였습니다"
			SendPacket(user, roomInNotice)
		} else if user.inRoom == true {
			user.curRoom.MsgChannel <- messagePacket
		}
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

	for {
		tryConnectClient, err := server.listener.Accept()

		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("새로운 클라이언트가 접속하였습니다")

		newClient := Client{iUserNum, false, false, Room{}, tryConnectClient}

		server.Users[newClient.netID] = newClient

		iUserNum += 1

		go UpdateUser(&newClient, &server)
	}
}

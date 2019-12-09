#include "pch.h"
#include "CChattingClient.h"
#include <iostream>

const char* ServerIP = "127.0.0.1";
const char* ServerPort = "9999";

CChattingClient::CChattingClient() {}

CChattingClient::~CChattingClient() {}

void CChattingClient::Init()
{
	if(WSAStartup(MAKEWORD(2, 2), &wsaData) != 0)
		CLog::ReportLog("WinSock DLL 로드 실패!", ELogType::LOG_ERROR);

	hSocket = WSASocket(AF_INET, SOCK_STREAM, 0, NULL, 0, WSA_FLAG_OVERLAPPED);

	if (hSocket == INVALID_SOCKET)
		CLog::ReportLog("소켓 생성 실패!", ELogType::LOG_ERROR);

	servAddr.sin_family = AF_INET;
	servAddr.sin_addr.s_addr = inet_addr(ServerIP);
	servAddr.sin_port = htons(atoi(ServerPort));

	if (connect(hSocket, (SOCKADDR*)&servAddr, sizeof(servAddr)) == SOCKET_ERROR)
		CLog::ReportLog("서버 연결 실패!", ELogType::LOG_ERROR);

	else
		CLog::ReportLog("서버 연결 성공", ELogType::LOG_DEFAULT);

	std::cout << "닉네임을 입력하세요 : ";
	std::cin >> nickname;

	std::cout << "===== 채팅서버에 입장하였습니다 =====" << std::endl;

	sendThread = std::thread([&]() { Send(); });
	readThread = std::thread([&]() { Read(); });
}

void CChattingClient::Update()
{
	while(true)
	{
		if (sendThread.joinable())
			sendThread.join();

		if (readThread.joinable())
			readThread.join();
	}
}

void CChattingClient::Destroy()
{
	WSACleanup();
}

void CChattingClient::Send()
{
	while (true)
	{
		DWORD sendBytes, dwFlags = 0;

		MESSAGE_PACKET* packet = new MESSAGE_PACKET;
		std::cin >> packet->message;

		char message[BUFSIZE] = "";
		sprintf(message, "[ %s ] : %s", nickname, packet->message);

		memset(&packet->overlapped, 0, sizeof(OVERLAPPED));
		packet->wsaBuffer.len = strlen(message) + 1;
		packet->wsaBuffer.buf = message;
		packet->bRead = false;

		WSASend(hSocket, &packet->wsaBuffer, 1, &sendBytes, dwFlags, &packet->overlapped, NULL);
	}
}

void CChattingClient::Read()
{
	while (true)
	{
		DWORD readBytes = 0;
		DWORD dwFlags = 0;

		WSAEVENT hEvent = WSACreateEvent();

		MESSAGE_PACKET* packet = new MESSAGE_PACKET;
		memset(&(packet->overlapped), 0, sizeof(OVERLAPPED));
		packet->overlapped.hEvent = hEvent;
		packet->wsaBuffer.len = BUFSIZE;
		packet->wsaBuffer.buf = packet->message;
		packet->bRead = true;

		WSARecv(hSocket,
			&(packet->wsaBuffer), 1,
			&readBytes, &dwFlags, &(packet->overlapped), NULL);

		WSAWaitForMultipleEvents(1, &hEvent, TRUE, WSA_INFINITE, FALSE);

		std::cout << packet->message << std::endl;
	}
}

#include "pch.h"
#include "CChattingServer.h"
#include <iostream>

#define ACTIVE_THREAD_NUM 3
#define MAX_USER 10

const char* ServerIP = "127.0.0.1";
const char* ServerPort = "9999";

CChattingServer::CChattingServer() {}
CChattingServer::~CChattingServer() {}

DWORD WINAPI CompletionThread(LPVOID CompletionPortIO);

void CChattingServer::Init()
{
	//WinSock 2.2 DLL 로드
	if (WSAStartup(MAKEWORD(2, 2), &wsaData) != 0)
		CLog::ReportLog("WinSock DLL 로드 실패!", ELogType::LOG_ERROR);

	//CompletionPort 생성
	completionPort = CreateIoCompletionPort(INVALID_HANDLE_VALUE, NULL, 0, ACTIVE_THREAD_NUM);

	//활성화하길 원하는 쓰레드 개수만큼 쓰레드 생성
	for (int i = 1; i <= ACTIVE_THREAD_NUM; i++)
	{
		switch (i)
		{
		case 1:
			thread1 = std::thread([&]() { CompletionThread((LPVOID)completionPort); });
			break;
		case 2:
			thread2 = std::thread([&]() { CompletionThread((LPVOID)completionPort); });
			break;
		case 3:
			thread3 = std::thread([&]() { CompletionThread((LPVOID)completionPort); });
			break;
		}
	}

	hServSock = WSASocket(AF_INET, SOCK_STREAM, 0, NULL, 0, WSA_FLAG_OVERLAPPED);

	if (hServSock == INVALID_SOCKET) 
		CLog::ReportLog("서버 소켓 생성 실패!", ELogType::LOG_ERROR);

	servAddr.sin_family = AF_INET;
	servAddr.sin_addr.s_addr = inet_addr(ServerIP);
	servAddr.sin_port = htons(atoi(ServerPort));

	if (bind(hServSock, (SOCKADDR*)&servAddr, sizeof(servAddr)) == SOCKET_ERROR)
		CLog::ReportLog("서버 소켓 등록 실패!", ELogType::LOG_ERROR);

	if (listen(hServSock, MAX_USER) == SOCKET_ERROR)
		CLog::ReportLog("서버가 클라이언트 연결 요청 대기 상태로 진입하지 못함!", ELogType::LOG_ERROR);

	else
		CLog::ReportLog("서버 구동 완료", ELogType::LOG_DEFAULT);
}

void CChattingServer::Update()
{
	while (true)
	{
		SOCKET hCIntSock;
		SOCKADDR_IN cIntAddr;
		int addrLen = sizeof(cIntAddr);

		hCIntSock = accept(hServSock, (SOCKADDR*)&cIntAddr, &addrLen);

		std::cout << "[ 클라이언트 연결 ] : " << hCIntSock << std::endl;

		handleData = new HANDLEDATA;
		handleData->hSocket = hCIntSock;
		memcpy(&(handleData->inSockAddr), &cIntAddr, addrLen);

		//Overlapped 소켓이랑 CompletionPort 연결
		CreateIoCompletionPort((HANDLE)hCIntSock, completionPort, (DWORD)handleData, 0);

		messagePacket = new MESSAGE_PACKET;
		memset(&(messagePacket->overlapped), 0, sizeof(OVERLAPPED));
		messagePacket->wsaBuffer.len = BUFSIZE;
		messagePacket->wsaBuffer.buf = messagePacket->message;
		Flags = 0;

		WSARecv(handleData->hSocket,
			&(messagePacket->wsaBuffer), 1,
			&recvBytes, &Flags, &(messagePacket->overlapped), NULL);
	}
}

void CChattingServer::Destroy()
{
	WSACleanup();
}

DWORD WINAPI CompletionThread(LPVOID CompletionPortIO)
{
	HANDLE hCompletionPort = (HANDLE)CompletionPortIO;
	DWORD BytesTransferred;
	HANDLEDATA* handleData;
	MESSAGE_PACKET* messagePacket;	
	DWORD flags;

	while (true)
	{
		//Input Output이 완료된 소켓 정보 얻음
		GetQueuedCompletionStatus(hCompletionPort,
			&BytesTransferred,
			(LPDWORD)&handleData,
			(LPOVERLAPPED*)&messagePacket,
			INFINITE);

		if (BytesTransferred == 0) //클라이언트가 연결 종료 요청
		{
			std::cout << "[ 클라이언트 종료 ] : " << handleData->hSocket << std::endl;
			closesocket(handleData->hSocket);
			delete handleData;
			delete messagePacket;
			continue;
		}

		messagePacket->wsaBuffer.len = BytesTransferred;
		WSASend(handleData->hSocket, &messagePacket->wsaBuffer, 1,
			NULL, 0, NULL, NULL);
	}
}
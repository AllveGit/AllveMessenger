// TCP_IP_Chattingserver.cpp : 이 파일에는 'main' 함수가 포함됩니다. 거기서 프로그램 실행이 시작되고 종료됩니다.
//

#include "pch.h"
#include <iostream>
#include <WinSock2.h>
#pragma comment(lib, "ws2_32.lib")

#include <thread>
#include <mutex>

#include "CLog.h"

#define BUFSIZE 100

typedef struct handleData
{
	SOCKET hSocket;
	SOCKADDR_IN inSockAddr;
} HANDLEDATA;

typedef struct messagePacket
{
	OVERLAPPED overlapped;
	char message[BUFSIZE];
	WSABUF wsaBuffer;
} MESSAGEPACKET;

int main()
{
	WSADATA wsaData;

	//WinSock 2.2 DLL 로드
	if (WSAStartup(MAKEWORD(2, 2), &wsaData) != 0)
	{
		CLog::ReportLog("WinSock DLL 로드 실패!", ELogType::LOG_ERROR);
	}



	WSACleanup();

	return 0;
}

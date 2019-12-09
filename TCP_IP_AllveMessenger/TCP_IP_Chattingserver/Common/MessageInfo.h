#pragma once

#include <WinSock2.h>
#pragma comment(lib, "ws2_32.lib")

#define BUFSIZE 100

struct HANDLEDATA
{
	SOCKET hSocket;
	SOCKADDR_IN inSockAddr;
};

struct MESSAGE_PACKET
{
	OVERLAPPED overlapped;
	char message[BUFSIZE];
	WSABUF wsaBuffer;
	bool bRead;
};
#pragma once
#include <thread>
#include <mutex>

#include "../Common/CLog.h"

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
};

class CChattingServer
{
private:
	WSADATA wsaData;
	HANDLE completionPort;
	SOCKADDR_IN servAddr;
	HANDLEDATA* handleData;
	MESSAGE_PACKET* messagePacket;
	SOCKET hServSock;
	DWORD recvBytes, Flags;
	int i;

	std::thread thread1, thread2, thread3;
public:
	CChattingServer();
	~CChattingServer();

	void Init();
	void Update();
	void Destroy();
};


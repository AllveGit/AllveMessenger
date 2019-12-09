#pragma once
#include <thread>
#include <mutex>
#include <list>

#include "../Common/CLog.h"
#include "../Common/MessageInfo.h"

class CChattingServer
{
private:
	WSADATA wsaData;
	HANDLE completionPort;
	SOCKADDR_IN servAddr;
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


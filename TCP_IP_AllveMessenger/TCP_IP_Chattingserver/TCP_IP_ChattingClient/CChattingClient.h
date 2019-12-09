#pragma once

#include <thread>

#include "../Common/CLog.h"
#include "../Common/MessageInfo.h"

class CChattingClient
{
private:
	WSADATA wsaData;
	SOCKET hSocket;
	SOCKADDR_IN servAddr;

	char nickname[20];

	std::thread sendThread;
	std::thread readThread;

	void Send();
	void Read();
public:
	CChattingClient();
	~CChattingClient();

	void Init();
	void Update();
	void Destroy();
};


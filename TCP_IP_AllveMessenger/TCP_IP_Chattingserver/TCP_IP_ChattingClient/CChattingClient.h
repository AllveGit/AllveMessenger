#pragma once
class CChattingClient
{
private:
	WSADATA wsaData;
	SOCKET hSocket;
	SOCKADDR_IN servAddr;
public:
	CChattingClient();
	~CChattingClient();

	void Init();
	void Update();
	void Destroy();
};


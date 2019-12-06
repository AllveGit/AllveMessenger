#include "pch.h"
#include "CChattingClient.h"

#include "CLog.h"

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

	//WSASend()
}

void CChattingClient::Update()
{
	while (true)
	{

	}
}

void CChattingClient::Destroy()
{
	WSACleanup();
}

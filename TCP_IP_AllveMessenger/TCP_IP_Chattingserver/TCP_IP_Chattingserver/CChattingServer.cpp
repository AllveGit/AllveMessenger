#include "pch.h"
#include "CChattingServer.h"
#include <iostream>

#define ACTIVE_THREAD_NUM 3
#define MAX_USER 10

std::list<HANDLEDATA*> userList;
std::mutex mtx_Lock;

const char* ServerIP = "127.0.0.1";
const char* ServerPort = "9999";

CChattingServer::CChattingServer() {}
CChattingServer::~CChattingServer() {}

DWORD WINAPI CompletionThread(LPVOID CompletionPortIO);

void CChattingServer::Init()
{
	//WinSock 2.2 DLL �ε�
	if (WSAStartup(MAKEWORD(2, 2), &wsaData) != 0)
		CLog::ReportLog("WinSock DLL �ε� ����!", ELogType::LOG_ERROR);

	//CompletionPort ����
	completionPort = CreateIoCompletionPort(INVALID_HANDLE_VALUE, NULL, 0, ACTIVE_THREAD_NUM);

	//Ȱ��ȭ�ϱ� ���ϴ� ������ ������ŭ ������ ����
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
		CLog::ReportLog("���� ���� ���� ����!", ELogType::LOG_ERROR);

	servAddr.sin_family = AF_INET;
	servAddr.sin_addr.s_addr = inet_addr(ServerIP);
	servAddr.sin_port = htons(atoi(ServerPort));

	if (bind(hServSock, (SOCKADDR*)&servAddr, sizeof(servAddr)) == SOCKET_ERROR)
		CLog::ReportLog("���� ���� ��� ����!", ELogType::LOG_ERROR);

	if (listen(hServSock, MAX_USER) == SOCKET_ERROR)
		CLog::ReportLog("������ Ŭ���̾�Ʈ ���� ��û ��� ���·� �������� ����!", ELogType::LOG_ERROR);

	else
		CLog::ReportLog("���� ���� �Ϸ�", ELogType::LOG_DEFAULT);
}

void CChattingServer::Update()
{
	while (true)
	{
		SOCKET hCIntSock;
		SOCKADDR_IN cIntAddr;
		int addrLen = sizeof(cIntAddr);

		hCIntSock = accept(hServSock, (SOCKADDR*)&cIntAddr, &addrLen);

		std::cout << "[ Ŭ���̾�Ʈ ���� ] : " << hCIntSock << std::endl;

		HANDLEDATA* handleData = new HANDLEDATA;
		handleData->hSocket = hCIntSock;
		memcpy(&(handleData->inSockAddr), &cIntAddr, addrLen);

		//Overlapped �����̶� CompletionPort ����
		CreateIoCompletionPort((HANDLE)hCIntSock, completionPort, (DWORD)handleData, 0);

		MESSAGE_PACKET* messagePacket = new MESSAGE_PACKET;
		memset(&(messagePacket->overlapped), 0, sizeof(OVERLAPPED));
		messagePacket->wsaBuffer.len = BUFSIZE;
		messagePacket->wsaBuffer.buf = messagePacket->message;
		Flags = 0;

		mtx_Lock.lock();
		userList.push_back(handleData);
		mtx_Lock.unlock();

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
	DWORD flags = 0;
	DWORD recvBytes = 0;

	while (true)
	{
		//Input Output�� �Ϸ�� ���� ���� ����
		GetQueuedCompletionStatus(hCompletionPort,
			&BytesTransferred,
			(LPDWORD)&handleData,
			(LPOVERLAPPED*)&messagePacket,
			INFINITE);

		if (BytesTransferred == 0) //Ŭ���̾�Ʈ�� ���� ���� ��û
		{
			mtx_Lock.lock();
			for (std::list<HANDLEDATA*>::iterator iter = userList.begin();
				iter != userList.end(); ++iter)
			{
				if ((*iter)->hSocket == handleData->hSocket)
				{
					userList.erase(iter);
					break;
				}
			}
			mtx_Lock.unlock();

			std::cout << "[ Ŭ���̾�Ʈ ���� ] : " << handleData->hSocket << std::endl;
			closesocket(handleData->hSocket);
			delete handleData;
			delete messagePacket;
			continue;
		}

		else if (messagePacket->bRead)
		{
			std::cout << messagePacket->message << std::endl;

			mtx_Lock.lock();
			for (auto iter : userList)
			{
				DWORD sendBytes = 0;
				flags = 0;

				MESSAGE_PACKET* sendPacket = new MESSAGE_PACKET;
				memset(&(sendPacket->overlapped), 0, sizeof(OVERLAPPED));
				sendPacket->wsaBuffer.len = BytesTransferred;
				sendPacket->wsaBuffer.buf = messagePacket->message;
				sendPacket->bRead = false;

				if (iter->hSocket != handleData->hSocket)
				{
					WSASend(iter->hSocket,
						&sendPacket->wsaBuffer, 1, &sendBytes, flags, &(sendPacket->overlapped), NULL);
				}
			}
			mtx_Lock.unlock();

			flags = 0;

			MESSAGE_PACKET* readPacket = new MESSAGE_PACKET;
			memset(&(readPacket->overlapped), 0, sizeof(OVERLAPPED));
			readPacket->wsaBuffer.len = BUFSIZE;
			readPacket->wsaBuffer.buf = readPacket->message;
			readPacket->bRead = true;

			WSARecv(handleData->hSocket,
				&(readPacket->wsaBuffer), 1,
				&recvBytes, &flags, &(readPacket->overlapped), NULL);
		}	
	}
}
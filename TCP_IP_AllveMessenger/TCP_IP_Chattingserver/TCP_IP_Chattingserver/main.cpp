// TCP_IP_Chattingserver.cpp : 이 파일에는 'main' 함수가 포함됩니다. 거기서 프로그램 실행이 시작되고 종료됩니다.
//

#include "pch.h"
#include <iostream>

#include "CChattingServer.h"

int main()
{
	CChattingServer server;

	server.Init();

	server.Update();

	server.Destroy();

	return 0;
}

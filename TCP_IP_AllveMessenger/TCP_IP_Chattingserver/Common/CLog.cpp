#include "pch.h"
#include "CLog.h"
#include <iostream>

void CLog::ReportLog(const char* LogMessage, ELogType _LogType)
{
	char LogTypeMessage[30] = "";
	bool bError = false;

	switch (_LogType)
	{
	case ELogType::LOG_DEFAULT:
		sprintf(LogTypeMessage, "%s", "LogType ( Default ) : ");
		break;
	case ELogType::LOG_WARNING:
		sprintf(LogTypeMessage, "%s", "LogType ( Warning ) : ");
		break;
	case ELogType::LOG_ERROR:
		sprintf(LogTypeMessage, "%s", "LogType ( Error ) : ");
		bError = true;
		break;
	default:
		sprintf(LogTypeMessage, "%s", "LogType ( Unknown ) : ");
		break;
	}
	
	std::cout << LogTypeMessage << LogMessage << std::endl;

	if (bError)
	{
		exit(1);
	}
}

#pragma once

enum ELogType
{
	LOG_DEFAULT = 0,
	LOG_WARNING,
	LOG_ERROR
};

class CLog
{
public:
	static void ReportLog(const char* LogMessage, ELogType _LogType);
};


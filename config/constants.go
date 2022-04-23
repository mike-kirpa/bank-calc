package config

const (
	ServiceName = "banks store service"
	//LogDebug has verbose message
	LogDebug = "debug"
	//LogInfo is default log level
	LogInfo = "info"
	//LogWarn is for logging messages about possible issues
	LogWarn = "warn"
	//LogError is for logging errors
	LogError = "error"
	//LogFatal is for logging fatal messages. The system shutdown after logging the message.
	LogFatal = "fatal"

	DetailFieldLockExpireAt = "lockExpireAt"
)

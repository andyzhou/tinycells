package logger

const (
	LogLevelOfDebug = "debug"
	LogLevelOfInfo 	= "info"
	LogLevelOfError = "error"
)

const (
	LogEnvOfLocal = "local"
	LogEnvOfProduction = "production"
)

const (
	DefaultLogPath = "./"
	DefaultLogFile = "logger.log"
	DefaultMaxSize = 100
	DefaultMaxBackups = 10
	DefaultMaxAge = 5
)
package logger

import "go.uber.org/zap"

type (
	//Logger = zap.Logger
	//Field = zap.Field
	//Sugar = zap.SugaredLogger
)

type (
	Config struct {
		Env string
		LogLevel string
		Rolling RollingConfig
		System  map[string]*RollingConfig
	}
	RollingConfig struct {
		Type       string
		FilePath   string
		FileName   string
		MaxSize    int
		MaxBackups int
		MaxAge     int
		Compress   bool
	}
)

var (
	Float32 = zap.Float32
	String 	= zap.String
	Any 	= zap.Any
	Int64 	= zap.Int64
	Int 	= zap.Int
	Int32 	= zap.Int32
	Uint 	= zap.Uint
	Duration = zap.Duration
	Durationp = zap.Durationp
	Object 	= zap.Object
	Namespace = zap.Namespace
	Reflect = zap.Reflect
	Skip 	= zap.Skip()
	ByteString = zap.ByteString
)
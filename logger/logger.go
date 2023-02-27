package logger

import (
	"errors"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"sync"
)

//global variable
var (
	_logger *Logger
	_loggerOnce sync.Once
)

//face info
type Logger struct {
	conf *Config
	logger *zap.Logger
	sysLogger map[string]*zap.Logger
	sync.RWMutex
}

//get single instance
func GetLogger() *Logger {
	_loggerOnce.Do(func() {
		_logger = NewLogger()
	})
	return _logger
}

//construct
func NewLogger(configs ...Config) *Logger {
	//self init
	this := &Logger{
		sysLogger: map[string]*zap.Logger{},
	}

	//check and setup config
	if configs != nil && len(configs) > 0 {
		//get config
		config := &configs[0]
		this.conf = config
		//inter init
		logger, err := this.initLogger(config.LogLevel, &config.Rolling)
		if err != nil {
			panic(err)
		}
		this.logger = logger
	}
	return this
}

//get default logger
func (f *Logger) S() *zap.Logger {
	if f.logger == nil {
		f.initDefaultLogger()
		if f.logger == nil {
			panic(errors.New("default logger hadn't init"))
		}
	}
	return f.logger
}

//get default sugared logger
func (f *Logger) SS() *zap.SugaredLogger {
	if f.logger == nil {
		f.initDefaultLogger()
		if f.logger == nil {
			panic(errors.New("default logger hadn't init"))
		}
	}
	return f.logger.Sugar()
}

//get sys logger
func (f *Logger) SysS(name string) *zap.SugaredLogger {
	v, ok := f.sysLogger[name]
	if ok && v != nil {
		//return matched
		return v.Sugar()
	}
	//return default
	if f.logger == nil {
		f.initDefaultLogger()
		if f.logger == nil {
			panic(errors.New("default logger hadn't init"))
		}
	}
	return f.logger.Sugar()
}

//set config
func (f *Logger) SetConfig(config *Config) error {
	//check
	if config == nil {
		return errors.New("invalid parameter")
	}
	f.conf = config

	//inter init
	logger, err := f.initLogger(config.LogLevel, &config.Rolling)
	if err != nil {
		return err
	}
	f.Lock()
	defer f.Unlock()
	f.logger = logger
	return nil
}

//build empty config
//build default config
func (f *Logger) BuildDefaultConfig() *Config {
	return &Config{
		Env: LogEnvOfLocal,
		LogLevel: LogLevelOfDebug,
		Rolling: RollingConfig{
			Type: LogEnvOfLocal,
			FilePath: DefaultLogPath,
			FileName: DefaultLogFile,
			MaxSize: DefaultMaxSize,
			MaxBackups: DefaultMaxBackups,
			MaxAge: DefaultMaxAge,
		},
	}
}

////////////////
//private func
////////////////

//init default logger
func (f *Logger) initDefaultLogger() {
	f.conf = f.BuildDefaultConfig()
	logger, err := f.initLogger(f.conf.LogLevel, &f.conf.Rolling)
	if err != nil || logger == nil {
		return
	}
	f.Lock()
	defer f.Unlock()
	f.logger = logger
}

//init sys logger
func (f *Logger) initSysLogger() bool {
	//check
	if f.conf == nil || len(f.conf.System) <= 0 {
		return false
	}
	for name, cfg := range f.conf.System {
		sysLogger, err := f.initLogger(f.conf.LogLevel, cfg)
		if err != nil || sysLogger == nil {
			continue
		}
		f.sysLogger[name] = sysLogger
	}
	return true
}

//filter level
func (f *Logger) levelFilter(level string, lvl zapcore.Level) bool {
	switch level {
	case LogLevelOfDebug:
		return lvl >= zapcore.DebugLevel
	case LogLevelOfInfo:
		return lvl >= zapcore.InfoLevel
	case LogLevelOfError:
		return lvl >= zapcore.ErrorLevel
	default:
		return lvl >= zapcore.DebugLevel
	}
}

//init logger
func (f *Logger) initLogger(level string, config interface{}) (*zap.Logger, error) {
	var (
		core zapcore.Core
	)

	//check
	rollingConfig, _ := config.(*RollingConfig)
	if level == "" {
		level = LogLevelOfDebug
	}
	if rollingConfig == nil {
		return nil, errors.New("invalid rolling config")
	}

	//inter init
	filePriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return f.levelFilter(level, lvl)
	})
	consolePriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return f.levelFilter(level, lvl)
	})
	fileWriteSync := zapcore.Lock(os.Stdout)
	productionConfig := zap.NewProductionEncoderConfig()
	productionConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewConsoleEncoder(productionConfig)

	//setup file write sync
	if rollingConfig.Type == LogEnvOfLocal {
		hook := lumberjack.Logger{
			Filename:   rollingConfig.FileName,
			MaxSize:    rollingConfig.MaxSize,
			MaxBackups: rollingConfig.MaxBackups,
			MaxAge:     rollingConfig.MaxAge,
			Compress:   rollingConfig.Compress,
		}
		fileWriteSync = zapcore.AddSync(&hook)
	}

	//setup console writer
	consoleWriteSync := zapcore.Lock(os.Stderr)
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

	//init zap core
	if f.conf.Env == LogEnvOfProduction {
		//for production env
		core = zapcore.NewTee(
			zapcore.NewCore(fileEncoder, fileWriteSync, filePriority),
		)
	}else{
		//for local env
		core = zapcore.NewTee(
			zapcore.NewCore(fileEncoder, fileWriteSync, filePriority),
			zapcore.NewCore(consoleEncoder, consoleWriteSync, consolePriority),
		)
	}
	return zap.New(core), nil
}


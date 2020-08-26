package log

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger

// errorsFile is the file where the errors are being written
var errorsFile *os.File

func init() {
	// default level: debug
	Init("debug", "")
}

// Init the logger with defined level. errorsPath defines the file where to store the errors, if set to "" will not store errors.
func Init(levelStr, errorsPath string) {
	var level zap.AtomicLevel
	err := level.UnmarshalText([]byte(levelStr))
	if err != nil {
		panic(fmt.Errorf("Error on setting log level: %s", err))
	}
	cfg := zap.Config{
		Level:            level,
		Encoding:         "console",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",

			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalColorLevelEncoder,

			TimeKey: "timestamp",
			EncodeTime: func(ts time.Time, encoder zapcore.PrimitiveArrayEncoder) {
				encoder.AppendString(ts.Local().Format(time.RFC3339))
			},
			EncodeDuration: zapcore.SecondsDurationEncoder,

			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,

			StacktraceKey: "stacktrace",
			LineEnding:    zapcore.DefaultLineEnding,
		},
	}

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	//nolint:errcheck
	defer logger.Sync()
	withOptions := logger.WithOptions(zap.AddCallerSkip(1))
	log = withOptions.Sugar()

	if errorsPath != "" {
		log.Infof("file where errors will be written: %s", errorsPath)
		errorsFile, err = os.OpenFile(errorsPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
	}

	log.Infof("log level: %s", level)
}

func writeToErrorsFile(msg string) {
	if errorsFile == nil {
		return
	}
	//nolint:errcheck
	errorsFile.WriteString(fmt.Sprintf("%s %s\n", time.Now().Format(time.RFC3339), msg))
}

// Debug calls log.Debug
func Debug(args ...interface{}) {
	log.Debug(args...)
}

// Info calls log.Info
func Info(args ...interface{}) {
	log.Info(args...)
}

// Warn calls log.Warn
func Warn(args ...interface{}) {
	log.Warn(args...)
}

// Error calls log.Error and stores the error message into the ErrorFile
func Error(args ...interface{}) {
	log.Error(args...)
	go writeToErrorsFile(fmt.Sprint(args...))
}

// Debugf calls log.Debugf
func Debugf(template string, args ...interface{}) {
	log.Debugf(template, args...)
}

// Infof calls log.Infof
func Infof(template string, args ...interface{}) {
	log.Infof(template, args...)
}

// Warnf calls log.Warnf
func Warnf(template string, args ...interface{}) {
	log.Warnf(template, args...)
}

// Errorf calls log.Errorf and stores the error message into the ErrorFile
func Errorf(template string, args ...interface{}) {
	log.Errorf(template, args...)
	go writeToErrorsFile(fmt.Sprintf(template, args...))
}

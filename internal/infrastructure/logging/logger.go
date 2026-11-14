package logging

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	*zap.SugaredLogger
	file *lumberjack.Logger
}

func New(env string) (*Logger, error) {

	logDir := "logs"
	logFile := filepath.Join(logDir, "app.log")

	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("Не удалось создать директорию для логов: %v\n", err.Error())
		os.Exit(1)
	}

	encoderConfig := zap.NewDevelopmentEncoderConfig()

	if env == "production" {
		encoderConfig = zap.NewProductionEncoderConfig()
	}
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	consoleConfig := encoderConfig
	consoleConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	fileConfig := encoderConfig
	fileConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	var consoleEncoder, fileEncoder zapcore.Encoder
	if env == "production" {
		consoleEncoder = zapcore.NewJSONEncoder(consoleConfig)
		fileEncoder = zapcore.NewJSONEncoder(fileConfig)
	} else {
		consoleEncoder = zapcore.NewConsoleEncoder(consoleConfig)
		fileEncoder = zapcore.NewConsoleEncoder(fileConfig)
	}

	// Open the file for recording
	logFileWriter := &lumberjack.Logger{Filename: logFile, MaxSize: 10, MaxBackups: 5, MaxAge: 30, Compress: true}

	level := zap.NewAtomicLevelAt(zapcore.DebugLevel)
	if env == "production" {
		level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), level)
	fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(logFileWriter), level)

	coreLogger := zapcore.NewTee(consoleCore, fileCore)

	zapLogger := zap.New(coreLogger, zap.WithCaller(false))

	return &Logger{
		SugaredLogger: zapLogger.Sugar(),
		file:          logFileWriter,
	}, nil
}

func (l *Logger) Close() error {
	if l == nil {
		return nil
	}
	var syncErr, closeErr error
	if l.SugaredLogger != nil {
		syncErr = l.SugaredLogger.Sync()
	}
	if l.file != nil {
		closeErr = l.file.Close()
	}
	return errors.Join(syncErr, closeErr)
}

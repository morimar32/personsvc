package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/ZachtimusPrime/Go-Splunk-HTTP/splunk"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	env "github.com/morimar32/helpers/environment"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	if err := env.LoadEnvironmentFile(); err != nil {
		log.Fatal(err)
	}
	initLogging()
	initDb()
}

func initDb() {
	connPart, err := env.GetEncryptedValue("connectionString")
	if err != nil {
		log.Fatal(err)
	}

	dbServer := env.GetValueWithDefault("DB_HOST", "localhost")
	connectionString = fmt.Sprintf(connPart, dbServer)
}

type SplunkWriter struct {
	Client  splunk.Client
	Writer  io.Writer
	entries chan []byte
}

func NewSplunkWriter(c splunk.Client) zapcore.WriteSyncer {
	writer := &SplunkWriter{}
	writer.Client = c
	sw := zapcore.Lock(writer)
	return sw
}
func (w *SplunkWriter) Sync() error {
	fmt.Println("SYNCING")
	return nil
}
func (w *SplunkWriter) Write(b []byte) (int, error) {
	w.Client.Log(string(b))
	return len(b), nil
}

func initLogging() {
	// lvl - global log level: Debug(-1), Info(0), Warn(1), Error(2), DPanic(3), Panic(4), Fatal(5)
	//globalLevel := zapcore.Level(0)
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return true
		//return lvl >= globalLevel && lvl < zapcore.ErrorLevel
	})
	consoleInfos := zapcore.Lock(os.Stdout)

	splunkClient := splunk.NewClient(
		nil,
		env.GetEncryptedValueWithDefault("splunkCollectorEndpoint", ""),
		env.GetEncryptedValueWithDefault("splunkHECToken", ""),
		env.GetValueWithDefault("splunkSource", ""),
		env.GetValueWithDefault("splunkSourceType", ""),
		env.GetValueWithDefault("splunkIndex", ""),
	)
	fmt.Println(env.GetEncryptedValueWithDefault("splunkCollectorEndpoint", ""))

	/*
		writer := &SplunkWriter{
			Writer: *w,
		}
	*/
	writer := NewSplunkWriter(*splunkClient)

	ecfg := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	consoleEncoder := zapcore.NewJSONEncoder(ecfg)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, writer, lowPriority),
		zapcore.NewCore(consoleEncoder, consoleInfos, lowPriority),
	)
	Log = zap.New(core)
	Log.Info("Logger created")
	zap.RedirectStdLog(Log)
	grpc_zap.ReplaceGrpcLogger(Log)

}

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"net"
	"net/http"
	"syscall"

	logging "personsvc/internal/logging"

	"github.com/ZachtimusPrime/Go-Splunk-HTTP/splunk"
	"github.com/cch123/gogctuner"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	env "github.com/morimar32/helpers/environment"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sys/unix"
)

func init() {
	if err := env.LoadEnvironmentFile(); err != nil {
		log.Fatal(err)
	}
	go gogctuner.NewTuner(true, 95)
	initLogging()
	initDb()
	initQueue()
}

func initDb() {
	connPart, err := env.GetEncryptedValue("connectionString")
	if err != nil {
		log.Fatal(err)
	}

	dbServer := env.GetValueWithDefault("DB_HOST", "localhost")
	connectionString = fmt.Sprintf(connPart, dbServer)
}

func initLogging() {
	// lvl - global log level: Debug(-1), Info(0), Warn(1), Error(2), DPanic(3), Panic(4), Fatal(5)
	logLevel, _ := strconv.Atoi(env.GetValueWithDefault("logLevel", "-1"))
	globalLevel := zapcore.Level(logLevel)
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		//return true
		return lvl >= globalLevel && lvl < zapcore.ErrorLevel
	})
	consoleInfos := zapcore.Lock(os.Stdout)

	dialer := &net.Dialer{
		Control: func(network, address string, conn syscall.RawConn) error {
			var operr error
			if err := conn.Control(func(fd uintptr) {
				operr = syscall.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.TCP_QUICKACK, 1) //linux
				//operr = syscall.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.TIOCGETA, 1) //osx
			}); err != nil {
				return err
			}
			return operr
		},
	}
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: dialer.DialContext,
		},
	}
	splunkClient := splunk.NewClient(
		client,
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
	writer := logging.NewSplunkWriter(*splunkClient)

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

func initQueue() {
	queueAddress = env.GetValueWithDefault("queueAddress", "")
	queueName = env.GetValueWithDefault("queueName", "")
}

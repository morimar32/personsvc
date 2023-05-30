package logging

import (
	"fmt"
	"io"

	"github.com/ZachtimusPrime/Go-Splunk-HTTP/splunk"
	"go.uber.org/zap/zapcore"
)

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

package service

import (
	"io"
	"os"
	"path/filepath"

	"monitor-agent/config"

	"github.com/sirupsen/logrus"
)

func InitLogger() {
	logPath := config.GetLogPath()
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}

	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logrus.SetOutput(io.MultiWriter(os.Stdout, file))
}

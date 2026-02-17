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
		logrus.SetOutput(os.Stdout)
		logrus.SetLevel(logrus.InfoLevel)
		return
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logrus.SetOutput(os.Stdout)
		logrus.SetLevel(logrus.InfoLevel)
		return
	}

	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logrus.SetOutput(io.MultiWriter(os.Stdout, file))
	logrus.SetLevel(logrus.InfoLevel)
	logrus.WithField("path", logPath).Info("Logger initialized")
}

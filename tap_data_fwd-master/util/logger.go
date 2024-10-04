package util

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LoggerManager struct {
	log        *logrus.Logger
	envManager *EnvironmentManager
}

func (LM *LoggerManager) InitLogger(envManager *EnvironmentManager) {
	LM.log = logrus.New()
	LM.envManager = envManager

	logDir := filepath.Join("/mnt/c/Users/deysh/OneDrive/Desktop/IRRA_FWD_Integrate/tap_data_fwd-master", "logs")

	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		err := os.MkdirAll(logDir, os.ModePerm)
		if err != nil {
			LM.log.Fatalf("Failed to create log directory: %v", err)
		}
	}

	logFilename := filepath.Join(logDir, "ULOG."+time.Now().Format("2006-01-02")+".log")

	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFilename,
		MaxSize:    50,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}

	logOutput := io.MultiWriter(lumberjackLogger, os.Stdout)
	LM.log.SetOutput(logOutput)

	LM.log.SetFormatter(&logrus.TextFormatter{
		TimestampFormat:        "2006-01-02 15:04:05",
		FullTimestamp:          true,
		ForceColors:            true,
		DisableColors:          false,
		DisableTimestamp:       false,
		DisableLevelTruncation: false,
		// DisableQuote:           true,
		// DisableTimestamp:       true,
	})

	LM.log.SetLevel(logrus.InfoLevel)
}

func (LM *LoggerManager) GetLogger() *logrus.Logger {
	return LM.log
}

func (LM *LoggerManager) LogInfo(serviceName string, msg string, args ...interface{}) {
	LM.log.Infof("[%s] "+msg, append([]interface{}{serviceName}, args...)...)
}

func (LM *LoggerManager) LogError(serviceName string, msg string, args ...interface{}) {
	LM.log.Errorf("[%s] "+msg, append([]interface{}{serviceName}, args...)...)
}

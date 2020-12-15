package rotatelog

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const logPath = "logs"

// InitLogFile init logrus with hook
func InitLogFile(module string, logs Logs) error {
	rotateTime, maxAge, err := logs.Durations()
	if err != nil {
		return err
	}

	if err := clearLockFiles(logPath); err != nil {
		return err
	}

	logrus.AddHook(newRotateHook(logPath, module, rotateTime, maxAge))
	logrus.SetOutput(ioutil.Discard)
	logLevel, err := logs.Level()
	if err != nil {
		logrus.WithField("error", err).Fatal("wrong log level")
	}

	logrus.SetLevel(logLevel)
	fmt.Printf("all logs are output in the %s directory, log level:%s\n", logPath, logs.LogLevel)
	return nil
}

func clearLockFiles(logPath string) error {
	files, err := ioutil.ReadDir(logPath)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	for _, file := range files {
		if ok := strings.HasSuffix(file.Name(), "_lock"); ok {
			if err := os.Remove(filepath.Join(logPath, file.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

// Logs logs cfg
type Logs struct {
	RotateTime string `json:"rotate_time"`
	MaxAge     string `json:"max_age"`
	LogLevel   string `json:"log_level"`
}

var defaultLogs = Logs{
	RotateTime: "24h",
	MaxAge:     "72h",
	LogLevel:   "debug",
}

// Durations return rotateTime, maxAge time.Duration
func (logs *Logs) Durations() (rotateTime, maxAge time.Duration, err error) {
	if logs.RotateTime == "" {
		logs.RotateTime = defaultLogs.RotateTime
	}

	if logs.MaxAge == "" {
		logs.MaxAge = defaultLogs.MaxAge
	}

	rotateTime, err = time.ParseDuration(logs.RotateTime)
	if err != nil {
		return
	}

	maxAge, err = time.ParseDuration(logs.MaxAge)
	if err != nil {
		return
	}

	return
}

// Level log level
func (logs *Logs) Level() (logrus.Level, error) {
	if logs.LogLevel == "" {
		logs.LogLevel = defaultLogs.LogLevel
	}

	return logrus.ParseLevel(logs.LogLevel)
}

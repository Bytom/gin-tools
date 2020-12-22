package rotatelog

import (
	"path/filepath"
	"sync"
	"time"

	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

var defaultFormatter = &logrus.TextFormatter{DisableColors: true}

type rotateHook struct {
	logPath      string
	module       string
	rotationTime time.Duration
	maxAge       time.Duration
	lock         *sync.Mutex
}

func newRotateHook(logPath, module string, rotationTime, maxAge time.Duration) *rotateHook {
	return &rotateHook{
		lock:         new(sync.Mutex),
		logPath:      logPath,
		module:       module,
		rotationTime: rotationTime,
		maxAge:       maxAge,
	}
}

// Write a log line to an io.Writer.
func (hook *rotateHook) ioWrite(entry *logrus.Entry) error {
	module := hook.module
	if data, ok := entry.Data["module"]; ok {
		module = data.(string)
	}

	logPath := filepath.Join(hook.logPath, module)
	writer, err := rotatelogs.New(
		logPath+".%Y%m%d%H%M%S",
		rotatelogs.WithMaxAge(hook.maxAge),
		rotatelogs.WithRotationTime(hook.rotationTime),
	)
	if err != nil {
		return err
	}

	msg, err := defaultFormatter.Format(entry)
	if err != nil {
		return err
	}

	if _, err = writer.Write(msg); err != nil {
		return err
	}

	return writer.Close()
}

// Fire write to file
func (hook *rotateHook) Fire(entry *logrus.Entry) error {
	hook.lock.Lock()
	defer hook.lock.Unlock()

	return hook.ioWrite(entry)
}

// Levels returns configured log levels.
func (hook *rotateHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

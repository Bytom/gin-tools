package rotatelog

import (
	"github.com/onrik/logrus/gorm"
	"github.com/sirupsen/logrus"
)

// GormLogger gorm logger
type GormLogger struct{}

// Print implement Logger interface
func (logger *GormLogger) Print(values ...interface{}) {
	logrus.Info(gorm.Formatter(values...))
}

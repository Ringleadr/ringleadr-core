package Logger

import (
	"errors"
	"fmt"
	"github.com/Ringleadr/ringleadr-core/internal/Utils"
	"go.uber.org/zap"
	"io"
	"log"
	"os"
)

var (
	logger  *zap.SugaredLogger
	out     io.Writer
	outName string
)

func InitLogger(background bool) {
	if background {
		file, err := setupLogFile()
		if err != nil {
			panic("Error setting up log directory: " + err.Error())
		}
		out = file
		outName = file.Name()
	} else {
		out = os.Stderr
		outName = "stderr"
	}

	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{
		outName,
	}
	baseLogger, err := config.Build()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	logger = baseLogger.Sugar()

	if background {
		logger.Infoln("Logging to file")
	} else {
		logger.Infoln("Logging to std out")
	}
}

func setupLogFile() (*os.File, error) {
	homeDir, err := Utils.GetUserHomeDir()
	if err != nil {
		return nil, err
	}
	dirName := fmt.Sprintf("%s/.agogos/host", homeDir)
	fileName := dirName + "/host.log"
	err = os.MkdirAll(dirName, 0755)
	if err != nil {
		return nil, errors.New("Could not make the Agogos log directory: " + err.Error())
	}
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, errors.New("Could not open/create a log file: " + err.Error())
	}
	return f, nil
}

func Logger() *zap.SugaredLogger {
	return logger
}

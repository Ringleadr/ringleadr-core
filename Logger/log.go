package Logger

import (
	"fmt"
	"github.com/GodlikePenguin/agogos-host/Utils"
	"github.com/pkg/errors"
	"io"
	"log"
	"os"
	"runtime/debug"
	"strings"
)

const logFlag int = log.Ldate | log.Ltime

var (
	logger    *log.Logger
	errLogger *log.Logger
	out       io.Writer
	errOut    io.Writer
)

func InitLogger(background bool) {
	if background {
		file, err := setupLogFile()
		if err != nil {
			panic("Error setting up log directory: " + err.Error())
		}
		out = file
		errOut = file
	} else {
		out = os.Stdout
		errOut = os.Stderr
	}
	logger = log.New(out, "[AGOGOS] ", logFlag)
	errLogger = log.New(errOut, "[AGOGOS] ", logFlag)
	if background {
		logger.Println("Logging to file")
	} else {
		logger.Println("Logging to std out")
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

func getCallerLine() string {
	stack := strings.Split(string(debug.Stack()), "\n")
	//ignore the first 7 items as they are the goroutine id and the call to debug.Stack, the call to the
	// Println/Printf function and the call to this function.
	//8th item is the function which called this one
	//9th item is the line of code that function call came from
	//remove tab from start of line and then return everything before the first space
	fullPath := strings.Split(strings.Replace(stack[8], "\t", "", 1), " ")[0]
	//Only get the file path after and including 'agogos-host'
	return fullPath[strings.LastIndex(fullPath, "agogos-host"):]
}

func Printf(format string, v ...interface{}) {
	v = append(v, getCallerLine())
	logger.Printf("["+format+"]\t%s", v...)
}

func Println(v ...interface{}) {
	logger.Println(v, "\t", getCallerLine())
}

func ErrPrintf(format string, v ...interface{}) {
	v = append(v, getCallerLine())
	errLogger.Printf("["+format+"]\t%s", v...)
}

func ErrPrintln(v ...interface{}) {
	errLogger.Println(v, "\t", getCallerLine())
}

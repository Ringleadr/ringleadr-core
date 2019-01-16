package Logger

import (
	"log"
	"os"
	"runtime/debug"
	"strings"
)

const logFlag int = log.Ldate | log.Ltime

var (
	logger    *log.Logger
	errLogger *log.Logger
)

func init() {
	logger = log.New(os.Stdout, "[AGOGOS] ", logFlag)
	errLogger = log.New(os.Stderr, "[AGOGOS] ", logFlag)
}

func getCallerLine() string {
	stack := strings.Split(string(debug.Stack()), "\n")
	//ignore the first 7 items as they are the goroutine id and the call to debug.Stack, the call to the
	// Println/Printf function and the call to this function.
	//8th item is the function which called this one
	//9th item is the line of code that function call came from
	//remove tab from start of line and then return everything before the first space
	return strings.Split(strings.Replace(stack[8], "\t", "", 1), " ")[0]
}

func Printf(format string, v ...interface{}) {
	v = append(v, getCallerLine())
	logger.Printf("["+format+"] %s", v...)
}

func Println(v ...interface{}) {
	logger.Println(v, getCallerLine())
}

func ErrPrintf(format string, v ...interface{}) {
	v = append(v, getCallerLine())
	errLogger.Printf("["+format+"] %s", v...)
}

func ErrPrintln(v ...interface{}) {
	errLogger.Println(v, getCallerLine())
}

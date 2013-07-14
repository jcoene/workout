package workout

import (
	"errors"
	"fmt"
	"github.com/jcoene/gologger"
	"strconv"
)

var log logger.Logger = *logger.NewLogger(logger.LOG_LEVEL_DEBUG, "workout")

func Error(f string, v ...interface{}) error {
	return errors.New(fmt.Sprintf(f, v...))
}

func parseInt(val string) (n int64) {
	n, _ = strconv.ParseInt(val, 10, 64)
	return
}

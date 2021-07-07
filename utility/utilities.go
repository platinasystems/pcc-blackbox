package utility

import (
	log "github.com/platinasystems/go-common/logs"
	"path/filepath"
	"runtime"
	"strings"
)

func FuncName(depth int) (string, string, int) {
	pc, file, line, _ := runtime.Caller(depth)
	nameFull := runtime.FuncForPC(pc).Name()
	log.AuctaLogger.Infof("Function name: %s", nameFull)
	nameEnd := filepath.Ext(nameFull)
	name := strings.TrimPrefix(nameEnd, ".")
	return name, file, line
}

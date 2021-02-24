package utility

import (
	"hash/fnv"
	"path/filepath"
	"runtime"
	"strings"

	log "github.com/platinasystems/go-common/logs"
)

func FuncName() string {
	pc, _, _, _ := runtime.Caller(2)
	nameFull := runtime.FuncForPC(pc).Name()
	log.AuctaLogger.Infof("Function name: %s", nameFull)
	nameEnd := filepath.Ext(nameFull)
	name := strings.TrimPrefix(nameEnd, ".")
	return name
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func GetTestID() uint32 {
	name := FuncName()
	return hash(name)
}

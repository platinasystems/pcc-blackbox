package utility

import (
	log "github.com/platinasystems/go-common/logs"
	"hash/fnv"
	"runtime"
)

func funcName() string {
	pc, _, _, _ := runtime.Caller(3)
	nameFull := runtime.FuncForPC(pc).Name()
	log.AuctaLogger.Infof("Function name: %s", nameFull)
	//nameEnd := filepath.Ext(nameFull)
	//name := strings.TrimPrefix(nameEnd, ".")
	return nameFull
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func GetTestID() uint32 {
	name := funcName()
	return hash(name)
}

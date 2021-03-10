package utility

import (
	log "github.com/platinasystems/go-common/logs"
	"math/rand"
	"time"
)

func CreateSeed() (seed int64) {
	seed = time.Now().Unix()
	log.AuctaLogger.Infof("Generated random seed: %d", seed)
	return
}

func RandomGenerator(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}

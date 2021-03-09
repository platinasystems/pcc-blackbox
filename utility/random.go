package utility

import (
	log "github.com/platinasystems/go-common/logs"
	"math/rand"
	"time"
)

func createSeed() (seed int64) {
	seed = time.Now().Unix()
	log.AuctaLogger.Infof("Random seed: %d", seed)
	return
}

func randomGenerator(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}

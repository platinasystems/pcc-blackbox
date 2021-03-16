// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"math/rand"
	"testing"
)

var runID string
var Env testEnv
var Pcc *pcc.PccClient
var Nodes = make(map[uint64]*pcc.NodeDetailed)
var SecurityKeys = make(map[string]*pcc.SecurityKey)
var NodebyHostIP = make(map[string]uint64) // deprecated use Env
var dockerStats *pcc.DockerStats
var randomGenerator *rand.Rand
var nameToTestFunc map[string]func(*testing.T)
var defaultTests map[string][]string

func main() {}

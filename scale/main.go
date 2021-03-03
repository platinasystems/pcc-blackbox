package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"

	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
)

type Config struct {
	count   int
	verbose bool
	search  string
	page    int
	limit   int
}

type actionFunc func()

var config Config
var envFile string = "env.json"
var containerFile string = "containers.json"
var env Env
var Pcc *pcc.PccClient
var dockerStats *pcc.DockerStats

const (
	ADDNODE     string = "addNode"
	DELNODE     string = "delNode"
	NODESUMMARY string = "nodeSummary"
	GETEVENT    string = "getEvent"

	VMTAG string = "vm"
)

func authenticate() {
	var err error

	credential := pcc.Credential{
		UserName: env.PccUser,
		Password: env.PccPassword,
	}
	if Pcc, err = pcc.Authenticate(env.PccIp, credential); err != nil {
		panic(fmt.Errorf("Authentication error: %v\n", err))
	}
}

func readEnvFile() {
	data, err := ioutil.ReadFile(envFile)
	if err != nil {
		panic(fmt.Errorf("Error opening %v: %v", envFile, err))
	}
	if err = json.Unmarshal(data, &env); err != nil {
		panic(fmt.Errorf("error unmarshalling %v: %v\n",
			envFile, err.Error()))
	}
	return
}

func main() {

	var (
		err    error
		action actionFunc
	)

	if len(os.Args) < 2 {
		panic("error: wrong number of arguments")
	}

	switch os.Args[1] {
	case ADDNODE:
		action = addNodeAction
		addNodeCmd := flag.NewFlagSet(ADDNODE, flag.ExitOnError)
		addNodeCmd.IntVar(&config.count, "n", 1, "count")
		addNodeCmd.BoolVar(&config.verbose, "v", false, "verbose")
		addNodeCmd.Parse(os.Args[2:])
	case DELNODE:
		action = delNodeAction
		delNodeCmd := flag.NewFlagSet(DELNODE, flag.ExitOnError)
		delNodeCmd.IntVar(&config.count, "n", math.MaxInt64, "count")
		delNodeCmd.BoolVar(&config.verbose, "v", false, "verbose")
		delNodeCmd.Parse(os.Args[2:])
	case NODESUMMARY:
		action = nodeSummaryAction
		nodeSummaryCmd := flag.NewFlagSet(NODESUMMARY, flag.ExitOnError)
		nodeSummaryCmd.BoolVar(&config.verbose, "v", false, "verbose")
		nodeSummaryCmd.Parse(os.Args[2:])
	case GETEVENT:
		action = getEventAction
		getEventCmd := flag.NewFlagSet(NODESUMMARY, flag.ExitOnError)
		getEventCmd.BoolVar(&config.verbose, "v", false, "verbose")
		getEventCmd.StringVar(&config.search, "s", "", "search")
		getEventCmd.IntVar(&config.page, "p", 0, "page")
		getEventCmd.IntVar(&config.limit, "l", 50, "limit")
		getEventCmd.Parse(os.Args[2:])
	default:
		panic("no action\n")
	}

	readEnvFile()

	err = pcc.StoreContainerNames()
	if err != nil {
		panic(fmt.Errorf("Error storing containers %v", err))
	}

	log.InitWithDefault(nil) // require to use pcc-client
	authenticate()

	dockerStats = pcc.InitDockerStats(env.DockerStats)
	action()
	dockerStats.Stop()
}

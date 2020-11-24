package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
)

type Config struct {
	count   int
	verbose bool
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

func storeContainerNames() (err error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return
	}

	// This assumes that this is running on the same CPU
	// as PCC blackbox.
	containers, err := cli.ContainerList(context.Background(),
		types.ContainerListOptions{})
	if err != nil {
		return
	}

	if len(containers) == 0 {
		return
	}

	m := make(map[string]string)
	for _, container := range containers {
		m[container.ID[:12]] = container.Names[0][1:]
	}

	data, err := json.MarshalIndent(m, "", "    ")
	if err != nil {
		err = fmt.Errorf("Error marshal to json: %v\n", err)
		return
	}
	err = ioutil.WriteFile(containerFile, data, 0644)
	return
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
	default:
		panic("no action\n")
	}

	readEnvFile()

	err = storeContainerNames()
	if err != nil {
		panic(fmt.Errorf("Error storing containers %v", err))
	}

	log.InitWithDefault(nil) // require to use pcc-client
	authenticate()

	dockerStats = pcc.InitDockerStats(env.DockerStats)
	action()
	dockerStats.Stop()
}

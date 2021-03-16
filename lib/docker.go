package pcc

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/KyleBanks/dockerstats"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	log "github.com/platinasystems/go-common/logs"
)

type DockerStatsConfig struct {
	OutputFile string
	Period     uint16
}

type DockerStats struct {
	fileName string
	phase    string
	start    *time.Time
	timer    *time.Timer
	file     *os.File
	writer   *bufio.Writer
}

// Init
func InitDockerStats(config DockerStatsConfig) *DockerStats {
	if config.OutputFile == "" {
		config.OutputFile = "container-stats.txt"
	}

	if config.Period <= 0 {
		config.Period = 30
	}

	var err error
	dockerStats := DockerStats{fileName: config.OutputFile}
	if dockerStats.file, err = os.Create(dockerStats.fileName); err == nil {
		dockerStats.writer = bufio.NewWriter(dockerStats.file)
		collect := func() {

			t := time.Now().Format(time.RFC3339)
			if stats, err := dockerstats.Current(); err == nil {
				for _, s := range stats {
					container := s.Container
					memory := s.Memory
					cpu := s.CPU
					io := s.IO
					pids := s.PIDs
					dockerStats.writer.WriteString(fmt.Sprintf("\n%s CONTAINER=%s: CPU=%v; MEMORY=%s; IO=%s; PIDS=%d", t, container, cpu, memory.String(), io.String(), pids))
					dockerStats.writer.Flush()
				}
			} else {
				log.AuctaLogger.Errorf("error collecting docker stats", err)
			}

			dockerStats.writer.Flush()
			dockerStats.timer.Reset(time.Second * time.Duration(config.Period)) // Write every 45s
		}
		dockerStats.timer = time.AfterFunc(time.Second*time.Duration(config.Period), collect) // start collect
	} else {
		panic(err)
	}

	return &dockerStats
}

// Switch test phase
func (ds *DockerStats) ChangePhase(name string) {
	if ds.phase != "" {
		start := ds.start
		end := time.Now()
		if ds.start != nil {
			_, _ = ds.writer.WriteString(fmt.Sprintf("\nEND %s; STARTTIME=%s; ENDTIME=%s; ELAPSEDTIME=%s", ds.phase, (*start).Format(time.RFC3339), end.Format(time.RFC3339), end.Sub(*start).String()))
		}
		ds.start = &end
		_, _ = ds.writer.WriteString(fmt.Sprintf("\n\nSTART %s; STARTTIME=%s", name, ds.start.Format(time.RFC3339)))
		ds.writer.Flush()
	}

	ds.phase = name
	ds.timer.Reset(time.Second * time.Duration(1))
}

// Stop
func (ds *DockerStats) Stop() {
	defer ds.file.Close()
	ds.writer.Flush()
	ds.timer.Stop()
}

func StoreContainerNames() (err error) {
	var containerFile string = "containers.json"

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

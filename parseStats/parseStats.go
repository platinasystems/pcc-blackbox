package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Stats struct {
	timeStamp   string
	container   string
	cpuPercent  string
	cpuPercentf float64
	mem1        string
	mem2        string
	memPercent  string
	memPercentf float64
	net1        string
	net2        string
	block1      string
	block2      string
	pids        string
}

var containers map[string]string
var statMap map[string][]Stats

var containerFile string = "containers.json"
var containerStatsFile string = "container-stats.txt"

const (
	STATTIME      int = 1
	STATCONTAINER int = 2
	STATCPUPERC   int = 3
	STATMEM1      int = 4
	STATMEM2      int = 5
	STATMEMPERC   int = 6
	STATNET1      int = 7
	STATNET2      int = 8
	STATIO1       int = 9
	STATIO2       int = 10
	STATIOPIDS    int = 11
	STATMAX       int = STATIOPIDS + 1
)

func readContainerNames() {

	data, err := ioutil.ReadFile(containerFile)
	if err != nil {
		panic(fmt.Errorf("Error opening %v: %v", containerFile, err))
	}
	if err := json.Unmarshal(data, &containers); err != nil {
		panic(fmt.Errorf("error unmarshalling %v: %v\n",
			containerFile, err.Error()))
	}

	statMap = make(map[string][]Stats)
	for k, _ := range containers {
		statMap[k] = []Stats{}
	}

	return
}

func parseLine(stat []string) {
	var (
		s   Stats
		err error
	)

	elements := len(stat)
	if elements != STATMAX {
		fmt.Printf("Error stat elements %d\n", elements)
		return
	}

	s.timeStamp = stat[STATTIME]
	s.container = stat[STATCONTAINER]
	s.cpuPercent = stat[STATCPUPERC]
	s.mem1 = stat[STATMEM1]
	s.mem2 = stat[STATMEM2]
	s.memPercent = stat[STATMEMPERC]
	s.net1 = stat[STATNET1]
	s.net2 = stat[STATNET2]
	s.block1 = stat[STATIO1]
	s.block2 = stat[STATIO2]
	s.pids = stat[STATIOPIDS]

	s.cpuPercentf, err = strconv.ParseFloat(s.cpuPercent, 64)
	if err != nil {
		fmt.Printf("Error converting memory cpuPercent: %v\n", err)
	}

	s.memPercentf, err = strconv.ParseFloat(s.memPercent, 64)
	if err != nil {
		fmt.Printf("Error converting memory percent: %v\n", err)
	}

	statMap[s.container] = append(statMap[s.container], s)
}

func readContainerStats() {
	file, err := os.Open(containerStatsFile)
	if err != nil {
		panic(fmt.Errorf("%v", err))
	}
	defer file.Close()

	var statsRegex = regexp.MustCompile(`^([0-9\-\:TZ]+)\s+CONTAINER=([a-f0-9]+):\s+CPU=([0-9\.]+)%;\s+MEMORY=Raw=([0-9\.TGMiB]+)\s+/\s+([0-9\.TGMiB]+)\s+Percent=([0-9\.]+)%;\s+IO=Network=([0-9\.TGMkB]+)\s+/\s+([0-9\.TGMkB]+)\s+Block=([0-9\.TGMB]+)\s+/\s+([0-9\.TGMB]+);\s+PIDS=([0-9]+)$`)

	scanner := bufio.NewScanner(file)
	var lines int = 0
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "START") {
			continue
		}
		lines++
		subMatch := statsRegex.FindAllStringSubmatch(line, -1)
		if len(subMatch) == 0 {
			fmt.Printf("Regex failed [%v]\n", line)
			continue
		}
		for _, elm := range subMatch {
			parseLine(elm)
		}
	}

	if err := scanner.Err(); err != nil {
		panic(fmt.Errorf("%v", err))
	}
}

func printStat(s Stats) {
	fmt.Printf("%v ", s.timeStamp)
	fmt.Printf("cpu %6s%% ", s.cpuPercent)
	fmt.Printf("memory %v/%v %v%% ", s.mem1, s.mem2, s.memPercent)
	fmt.Printf("net %v/%v ", s.net1, s.net2)
	fmt.Printf("block %v/%v ", s.block1, s.block2)
	fmt.Printf("pids %v\n", s.pids)
}

func printStats() {
	for k, v := range statMap {
		if name, ok := containers[k]; ok {
			fmt.Printf("\nContainer [%v]\n", name)
		} else {
			fmt.Printf("\nContainer [%v]\n", k)
		}
		for _, s := range v {
			printStat(s)
		}
	}
}

type statSummary struct {
	container string
	maxCpu    float64
	avgCpu    float64
	maxMem    float64
	avgMem    float64
}

func (s statSummary) String() string {
	return fmt.Sprintf("%-19s max cpu %6.2f%% avg cpu %6.2f%%,"+
		"\tmax mem %6.2f%% avg mem %6.2f%%",
		containers[s.container], s.maxCpu, s.avgCpu, s.maxMem, s.avgMem)
}

func analyzeTop() {
	var (
		sum           statSummary
		sumContainers []statSummary
	)
	for k, v := range statMap {
		cpuCount := 0
		memCount := 0
		sum.container = k
		sum.maxCpu = 0.0
		sum.avgCpu = 0.0
		sum.maxMem = 0.0
		sum.avgMem = 0.0
		for _, s := range v {
			sum.avgCpu += s.cpuPercentf
			sum.avgMem += s.memPercentf
			if s.cpuPercentf > sum.maxCpu {
				sum.maxCpu = s.cpuPercentf
			}
			if s.memPercentf > sum.maxMem {
				sum.maxMem = s.memPercentf
			}
			cpuCount++
			memCount++
		}
		if cpuCount > 0 {
			sum.avgCpu = sum.avgCpu / float64(cpuCount)
		}
		if memCount > 0 {
			sum.avgMem = sum.avgMem / float64(memCount)
		}
		sumContainers = append(sumContainers, sum)
	}

	// sort by maxCpu
	sort.Slice(sumContainers, func(i, j int) bool {
		return sumContainers[i].maxCpu > sumContainers[j].maxCpu
	})
	fmt.Printf("\nSort by max cpu:\n================\n")
	for _, s := range sumContainers {
		fmt.Printf("%v\n", s)

	}

	// sort by maxMem
	sort.Slice(sumContainers, func(i, j int) bool {
		return sumContainers[i].maxMem > sumContainers[j].maxMem
	})
	fmt.Printf("\nSort by max memory:\n===================\n")
	for _, s := range sumContainers {
		fmt.Printf("%v\n", s)
	}

}

func parseStats() {

	readContainerNames()
	readContainerStats()
	printStats()
	analyzeTop()

	return
}

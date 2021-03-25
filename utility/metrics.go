package utility

import (
	"encoding/json"
	"fmt"
	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"io/ioutil"
	"os"
)

func SaveTopicDataOnFiles(runID string, data *map[string]interface{}, topic string) {
	path := fmt.Sprintf("nodes_monitoring/%s", runID)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			log.AuctaLogger.Error("Cannot create directories %s", path)
			return
		}
	}
	for k, v := range *data {
		fileName := fmt.Sprintf("%s_%s", k, topic)
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			log.AuctaLogger.Errorf("%v", err)
			return
		}
		err = ioutil.WriteFile(fmt.Sprintf("%s/%s", path, fileName), b, 0644)
		if err != nil {
			log.AuctaLogger.Errorf("%v", err)
			return
		}
	}
}

func SaveNodesHistoricalSummaries(pcc *pcc.PccClient, runID string, startTime uint64, stopTime uint64) {
	/*fields := []string {"cpuLoad",
	"availableMem",
	"realUsedMem",
	"diskUsage",
	"inodeUsage",
	"cpuTemp",
	"networkThroughput"}*/
	fields := make([]string, 0)
	topic := "summary"
	nodes, err := pcc.GetNodes()
	if err != nil {
		log.AuctaLogger.Errorf("%v", err)
		return
	}
	var nodeIDs []uint64
	for _, node := range *nodes {
		nodeIDs = append(nodeIDs, node.Id)
	}

	if data, err := pcc.GetNodesHistory(topic, startTime, stopTime, nodeIDs, fields); err != nil {
		log.AuctaLogger.Errorf("Error in getting history for topic %s, %v", topic, err)
		return
	} else {
		if len(*data) == 0 {
			log.AuctaLogger.Infof("No data for topic summary in the timeRange: %d - %d", startTime, stopTime)
			return
		} else {
			SaveTopicDataOnFiles(runID, data, topic)
		}
	}
}

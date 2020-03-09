package main

import (
	"fmt"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
	"testing"
	"time"
)

func monitorTest(t *testing.T) {
	t.Run("websocketTest", monitorWebsocketTest)
	t.Run("monitorRestTest", monitorRestTest)
}

func monitorWebsocketTest(t *testing.T) {

	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	// Monitor WS
	var ws *pcc.Ws = Pcc.WsLiveDataConnect()
	if ws.Error != nil {
		assert.Fatalf("Websocket connection failed: %+v\n", ws.Error)
	}
	defer ws.Connection.Close()

	// Start go routine in order to listen live data
	ws.WsListen()

	// Send command
	ws.WsSendingMetrics([]string{"cpu", "memory", "memory", "sensor", "network", "nodeDetails-2", "podDetails-2", "svcDetails-2", "storage"})

	// Wait listen timeout
	ws.WsListenTimeout(30 * time.Second)

	for id, node := range Nodes {
		if Pcc.IsNodeOnline(node) {
			n, ok := ws.Statistics.ActiveNodeMap[id]
			if !ok {
				assert.Fatalf("Unable to get data from node %d", id)
			} else {
				fmt.Printf("Node [%d] received '%d' messages from Websocket\n", id, n)
			}
		}
	}
}

func monitorRestTest(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	from := time.Now().Add(-time.Hour).Unix() * 1000
	to := time.Now().Unix() * 1000

	fromT := time.Unix(from/1000, 0)
	toT := time.Unix(to/1000, 0)

	fmt.Printf("Time window: from %s (%d) -> to %s (%d) \n", fromT, from, toT, to)

	for id, node := range Nodes {
		if Pcc.IsNodeOnline(node) {

			// CPU
			fmt.Printf("Executing CPU REST on node [%d] ", id)
			resultCpu, err := Pcc.GetHistoricalData("cpu", from, to, []uint64{id}, []string{"us", "sy", "id"})
			if err != nil {
				assert.Fatalf("Unable to get historical data from node [%d]: Error is %s", id, err)
			} else if resultCpu == "" {
				assert.Fatalf("No CPU data has been collected from '%s'to '%s'", fromT, toT)
			} else {
				fmt.Printf("OK\n")
			}

			// MEMORY
			fmt.Printf("Execute MEMORY REST on node [%d] ", id)
			resultMemory, err := Pcc.GetHistoricalData("memory", from, to, []uint64{id}, []string{"us", "sy", "id"})
			if err != nil {
			} else if resultMemory == "" {
				assert.Fatalf("No MEMORY data has been collected from '%s'to '%s'", fromT, toT)
			} else {
				fmt.Printf("OK\n")
			}

			// DISK
			fmt.Printf("Execute DISK REST on node [%d] ", id)
			resultDisk, err := Pcc.GetHistoricalData("disk", from, to, []uint64{id}, []string{"us", "sy", "id"})
			if err != nil {
			} else if resultDisk == "" {
				assert.Fatalf("No DISK data has been collected from '%s'to '%s'", fromT, toT)
			} else {
				fmt.Printf("OK\n")
			}
		}
	}
}

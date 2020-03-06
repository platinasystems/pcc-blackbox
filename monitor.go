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

	// Get token
	token := Pcc.GetBearer()

	// Monitor WS
	m := pcc.NewMonitor(token, fmt.Sprintf("%s:9999", Env.PccIp), true)
	if !m.WSConnect() {
		assert.Fatalf("Connection failed: %+v\n", m.Ws.Error)
	}
	defer m.Ws.Connection.Close()

	m.WsListeningLoop()
	m.WsSendingMetrics([]string{"cpu", "memory", "memory", "sensor", "network", "nodeDetails-2", "podDetails-2", "svcDetails-2", "storage"})
	m.WsReadTimeout(30 * time.Second)

	for id, node := range Nodes {
		if Pcc.IsNodeOnline(node) {
			n, ok := m.Ws.Statistics.ActiveNodeMap[id]
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

	// Get token
	token := Pcc.GetBearer()

	m := pcc.NewMonitor(token, fmt.Sprintf("%s:9999", Env.PccIp), true)

	from := time.Now().Add(-time.Hour).Unix() * 1000
	to := time.Now().Unix() * 1000

	fromT := time.Unix(from/1000, 0)
	toT := time.Unix(to/1000, 0)

	fmt.Printf("Time window: from %s (%d) -> to %s (%d) \n", fromT, from, toT, to)

	for id, node := range Nodes {
		if Pcc.IsNodeOnline(node) {

			// CPU
			fmt.Printf("Executing CPU REST on node [%d] ", id)
			resultCpu, err := m.GetHistorical("cpu", from, to, []uint64{id}, []string{"us", "sy", "id"})
			if err != nil {
				assert.Fatalf("Unable to get historical data from node [%d]: Error is %s", id, err)
			} else if resultCpu == "" {
				assert.Fatalf("No CPU data has been collected from '%s'to '%s'", fromT, toT)
			} else {
				fmt.Printf("OK\n")
			}

			// MEMORY
			fmt.Printf("Execute MEMORY REST on node [%d] ", id)
			resultMemory, err := m.GetHistorical("memory", from, to, []uint64{id}, []string{"us", "sy", "id"})
			if err != nil {
			} else if resultMemory == "" {
				assert.Fatalf("No MEMORY data has been collected from '%s'to '%s'", fromT, toT)
			} else {
				fmt.Printf("OK\n")
			}

			// DISK
			fmt.Printf("Execute DISK REST on node [%d] ", id)
			resultDisk, err := m.GetHistorical("disk", from, to, []uint64{id}, []string{"us", "sy", "id"})
			if err != nil {
			} else if resultDisk == "" {
				assert.Fatalf("No DISK data has been collected from '%s'to '%s'", fromT, toT)
			} else {
				fmt.Printf("OK\n")
			}
		}
	}
}

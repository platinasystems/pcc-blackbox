package main

import (
	"fmt"
	"time"

	pcc "github.com/platinasystems/pcc-blackbox/lib"
)

type testNode struct {
	ip      string
	id      uint64
	node    pcc.NodeDetailed
	start   time.Time
	elapsed time.Duration
	done    bool
}

var nodes []testNode

func addNodeAction() {
	dockerStats.ChangePhase("addNodes")
	addNodes()
	dockerStats.ChangePhase("waitNodes")
	waitNodes()
}

func addDone() {

	var totalElapsed time.Duration
	var count int64

	fmt.Printf("Summary\n")
	fmt.Printf("=======\n")
	for _, n := range nodes {
		count++
		pStatus, err := Pcc.GetProvisionStatus(n.id)
		if err != nil {
			fmt.Printf("Error GetProvisionStatus: %v\n",
				err)
			continue
		}
		cStatus, err := Pcc.GetNodeConnectionStatus(n.id)
		if err != nil {
			fmt.Printf("Error GetNodeProvisionStatus: %v\n",
				err)
			continue
		}
		fmt.Printf("node %3v %v provision [%v] connection [%v] "+
			"elapsed %v\n",
			n.id, n.ip, pStatus, cStatus, n.elapsed)
		totalElapsed += n.elapsed
	}
	if count > 0 {
		avg := int64(totalElapsed) / count
		fmt.Printf("\nAverage for %d nodes - %v\n",
			count, time.Duration(avg))
	}
}

func addNodes() {

	var (
		err       error
		nodeCount int
	)

	for _, n := range env.Nodes {
		var newNode testNode

		if _, err = Pcc.FindNodeAddress(n); err == nil {
			if config.verbose {
				fmt.Printf("node %v already exists\n", n)
			}
			continue
		}
		newNode.node.Managed = new(bool)
		*newNode.node.Managed = true
		newNode.node.Host = n
		newNode.node.Tags = []string{VMTAG}
		newNode.ip = n
		newNode.done = false

		if config.verbose {
			fmt.Printf("Adding [%v]\t", n)
		}
		newNode.start = time.Now()
		err = Pcc.AddNode(&newNode.node)
		if err != nil {
			fmt.Printf("\nError adding [%v]: %v\n", n, err)
		}
		newNode.id = newNode.node.Id
		if config.verbose {
			fmt.Printf("node %v\n", newNode.id)
		}

		nodes = append(nodes, newNode)
		nodeCount++
		if nodeCount == config.count {
			break
		}
	}
}

func waitNodes() {

	var (
		done      int
		nodeCount int
	)

	done = 0
	nodeCount = len(nodes)
	for {
		for i, n := range nodes {
			if n.done {
				done++
				continue
			}
			pStatus, err := Pcc.GetProvisionStatus(n.id)
			if err != nil {
				fmt.Printf("Error GetProvisionStatus: %v\n",
					err)
				continue
			}
			cStatus, err := Pcc.GetNodeConnectionStatus(n.id)
			if err != nil {
				fmt.Printf("Error GetNodeProvisionStatus: %v\n",
					err)
				continue
			}
			if config.verbose {
				fmt.Printf("%v p [%v] c [%v] elapsed %v\n",
					n.ip, pStatus, cStatus,
					time.Since(n.start))
			}
			statuses, err := Pcc.GetStatusId(n.id, pcc.RUNNING)
			if err != nil {
				fmt.Printf("Error GetStatusId: %v\n", err)
				continue

			}
			oks, err := Pcc.GetStatusId(n.id, pcc.OK)
			if err != nil {
				fmt.Printf("Error GetStatusId: %v\n", err)
				continue
			}
			numStatuses := len(statuses)
			numOks := len(oks)
			if config.verbose {
				fmt.Printf("running %d ok %d\n",
					numStatuses, numOks)
			}
			if false && config.verbose && numStatuses > 0 {
				for _, s := range statuses {
					fmt.Printf("\t%+v\n", s)
				}
			}
			if pStatus != "Adding node..." && cStatus == "online" {
				if numStatuses == 0 && n.done == false {
					n.done = true
					done++
					n.elapsed = time.Since(n.start)
					nodes[i] = n
				}
			}

		}
		if config.verbose {
			fmt.Println("")
		}
		if done == nodeCount {
			addDone()
			return
		}
		done = 0
		time.Sleep(10 * time.Second)
	}
}

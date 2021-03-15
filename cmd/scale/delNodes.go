package main

import (
	"fmt"
	"strings"
	"time"
)

func delNodeAction() {
	dockerStats.ChangePhase("delNodes")
	delNodes()
	dockerStats.ChangePhase("waitDelNodes")
	waitDelNodes()
}

func delNodes() {
	var (
		err       error
		newNode   testNode
		nodeCount int
	)

	currentNodes, err := Pcc.GetNodes()
	if err != nil {
		panic(fmt.Errorf("failed to GetNodes: %v\n", err))
	}
	for _, n := range *currentNodes {
		if config.verbose {
			fmt.Printf("checking node %v [%v]\n", n.Id, n.Host)
		}
		for _, t := range n.Tags {
			if t == VMTAG {
				if config.verbose {
					fmt.Printf("deleting %v\n", n.Id)
				}
				newNode.id = n.Id
				newNode.ip = n.Host
				Pcc.DeleteNode(n.Id)
				newNode.start = time.Now()
				nodes = append(nodes, newNode)
				nodeCount++
			}
		}
		if nodeCount == config.count {
			break
		}
	}
}

func waitDelNodes() {
	var done bool
	for !done {
		done = true
		for i, n := range nodes {
			if n.done {
				continue
			}
			if config.verbose {
				fmt.Printf("checking %v - ", n.id)
			}
			node, err := Pcc.GetNode(n.id)
			if err != nil {
				if strings.Contains(err.Error(),
					"record not found") {
					n.done = true
					n.elapsed = time.Since(n.start)
					nodes[i] = n
					continue
				}
				fmt.Printf("GetNode err [%v]\n", err)
			}
			done = false
			if config.verbose {
				fmt.Printf("%v\n", node.ProvisionStatus)
			}
		}
		time.Sleep(time.Second * 10)
	}
	delDone()
}

func delDone() {

	fmt.Printf("Summary\n")
	fmt.Printf("=======\n")
	for _, n := range nodes {
		fmt.Printf("node %v %v deleted, elapsed %v\n",
			n.id, n.ip, n.elapsed)
	}
}

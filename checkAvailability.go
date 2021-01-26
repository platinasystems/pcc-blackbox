package main

import (
	"fmt"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"strings"
	"sync"
	"testing"
	"time"
)

// Add UNREACHABLE NODE
func addUnreachableNode(t *testing.T) {
	address := Env.Availability.FakeAddress
	if address == "" {
		address = "IamNotWorking"
	}
	if err := checkNodeConnectionStatus("unreachable", address); err != nil {
		t.Fatal(err)
	}
}

// Add INACCESSIBLE NODE
func addInaccessibleNode(t *testing.T) {
	address := Env.Availability.Inaccessible
	if address == "" {
		address = Env.PccIp
	}
	if err := checkNodeConnectionStatus("inaccessible", address); err != nil {
		t.Fatal(err)
	}
}

// add a node and wait for the connection status
func checkNodeConnectionStatus(status string, host string) (err error) {
	fmt.Println(fmt.Sprintf("\nAVAILABILITY: add %s node", status))
	var node *pcc.NodeDetailed

	node = &pcc.NodeDetailed{}
	node.Host = host

	if err = Pcc.AddNode(node); err == nil {
		defer func() { // delete the node
			Pcc.DeleteNode(node.Id)
		}()

		fmt.Printf("node [%s] added with id [%d]. Waiting for connection status [%s]\n", node.Host, node.Id, status)
		for i := 1; i <= 20; i++ { // wait for the status
			time.Sleep(time.Second * time.Duration(10))
			if node, err = Pcc.GetNode(node.Id); err == nil && node.NodeStatus != nil {
				connectionStatus := node.NodeStatus.ConnectionStatus
				fmt.Printf("Connection status for node %s is %s\n", host, connectionStatus)
				if strings.Compare(strings.ToLower(connectionStatus), status) == 0 {
					return
				}
			}
		}
	}

	if err == nil {
		err = fmt.Errorf("unable to get the fake status for the node %s", node.Host)
	}

	return
}

// Delete  both agent and collector and wait for the restore
func checkAgentAndCollectorRestore(t *testing.T) {
	fmt.Println("\nAVAILABILITY: checking the agent/collector restore")
	var wg sync.WaitGroup

	if nodes, err := Pcc.GetNodes(); err == nil {
		if len(*nodes) > 0 {
			node := (*nodes)[0] // get the first node
			wg.Add(2)

			f := func(service string, path string) {
				defer wg.Done()
				if err = checkRestore(service, path, &node); err != nil {
					t.Fatal(err)
				}
			}

			go f(AGENT_NOTIFICATION, "pccagent")            // Delete the agent and wait for the restore
			go f(COLLECTOR_NOTIFICATION, "systemCollector") // Delete the collector and wait for the restore

			wg.Wait()
		} else {
			t.Fatal("No nodes were found")
		}
	} else {
		t.Fatal(err)
	}
}

// Remove the service and wait for the restore
func checkRestore(service string, path string, node *pcc.NodeDetailed) (err error) {
	var (
		ssh      pcc.SSHHandler
		execPath string
		stdout   string
		cmd      string
	)
	execPath = fmt.Sprintf("/opt/platina/pcc/bin/%v", path)

	fmt.Printf("Stopping and removing the service %s from node %d %s %s\n",
		service, node.Id, node.Name, node.Host)
	cmd = fmt.Sprintf("sudo rm -f %s && sudo kill -9 `pidof %s`",
		execPath, path)
	if _, _, err = ssh.Run(node.Host, cmd); err != nil {
		return
	}
	fmt.Printf("The %s:%s was correctly killed & removed from node %d:%s\n",
		service, path, node.Id, node.Name)

	time.Sleep(10 * time.Second)
	timeout := time.After(10 * time.Minute)
	tick := time.Tick(10 * time.Second)
	done := false
	cmd = fmt.Sprintf("if [ -f '%s' ]; then echo -n OK; fi", execPath)
	for !done {
		select {
		case <-timeout:
			err = fmt.Errorf("Timed out waiting for %v\n",
				service)
			return
		case <-tick:
			stdout, _, err = ssh.Run(node.Host, cmd)
			if err != nil {
				return
			}
			if stdout == "OK" {
				fmt.Printf("Executable found: %s\n", path)
				done = true
			}
		}
	}

	cmd = fmt.Sprintf("pidof %s", path)
	stdout, _, err = ssh.Run(node.Host, cmd)
	if err != nil {
		err = fmt.Errorf("Could not find pid of %s\n", path)
		return
	}
	fmt.Printf("pid of %s found %s", path, stdout)

	return
}

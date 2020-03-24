package main

import (
	"fmt"
	"os/exec"
	"sync"
	"testing"
	"time"
)

// Delete servers and pxeboot (in parallel)
func addGreenfieldServers(t *testing.T) {
	fmt.Println("\nGREENFIELD: executing the pxeboot for the servers")
	var wg sync.WaitGroup

	if err := Pcc.DeleteServers(true); err != nil { // wait for the deletion
		t.Fatal(err)
	}

	wg.Add(len(Env.Servers))

	pxeboot := func(server node) {
		defer wg.Done()
		bmc := server.BMCIp
		host := server.HostIp
		fmt.Println(fmt.Sprintf("executing the pxeboot for server [%s]", bmc))
		cmd := exec.Command("ipmitool", "-I", "lanplus", "-H", bmc, "-U", "ADMIN", "-P", "ADMIN", "chassis", "bootdev", "pxe")
		if err := cmd.Run(); err == nil {
			cmd = exec.Command("ipmitool", "-I", "lanplus", "-H", bmc, "-U", "ADMIN", "-P", "ADMIN", "chassis", "power", "cycle")
			if err = cmd.Run(); err == nil {

				for i := 1; i <= 20; i++ { //wait for the node
					time.Sleep(time.Second * 15)
					if nodes, err := Pcc.GetNodes(); err == nil {
						for _, node := range *nodes {
							if node.Bmc == bmc {
								fmt.Println(fmt.Sprintf("the pxeboot for the server [%s] was completed. node added with id [%d]", bmc, node.Id))
								node.Host = host
								server.Id = node.Id
								if err = Pcc.UpdateNode(&node); err != nil { // Set the host address for the greenfield server
									t.Fatal(err)
								}
								return
							}
						}
					}
				}
			} else {
				t.Fatal(err)
			}
		} else {
			t.Fatal(err)
		}
	}

	for _, server := range Env.Servers { // Start the pxeboot
		go pxeboot(server.node)
	}

	fmt.Println("PXEBOOT: wait for the servers")

	wg.Wait()

}

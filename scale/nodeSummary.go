package main

import (
	"fmt"
	"time"
)

var totalElapsed time.Duration

func nodeSummaryAction() {
	dockerStats.ChangePhase("getTenants")
	getTenants()
	dockerStats.ChangePhase("getNodeGroups")
	getNodeGroups()
	dockerStats.ChangePhase("getRoles")
	getRoles()
	dockerStats.ChangePhase("getTemplates")
	getTemplates()
	//dockerStats.ChangePhase("getNodes")
	// getNodes()
	dockerStats.ChangePhase("getNodesOptions")
	getNodesOptions()

	fmt.Printf("\n%-22s = %v\n", "Total elapsed", totalElapsed)
}

func getTenants() {
	start := time.Now()
	tenants, err := Pcc.GetTenants()
	elapsed := time.Since(start)
	if err != nil {
		fmt.Printf("Error GetTenants: %v\n", err)
		return
	}
	if config.verbose {
		fmt.Println("")
	}
	fmt.Printf("%-18s %3d = %v\n", "GetTenants", len(tenants), elapsed)
	for i, t := range tenants {
		if config.verbose {
			fmt.Printf("\t%d tenant [%v]\n", i, t.Name)
		}
	}
	totalElapsed += elapsed
}

func getNodeGroups() {
	start := time.Now()
	groups, err := Pcc.GetNodeGroups()
	elapsed := time.Since(start)
	if err != nil {
		fmt.Printf("Error GetNodeGroups: %v\n", err)
		return
	}
	if config.verbose {
		fmt.Println("")
	}
	fmt.Printf("%-18s %3d = %v\n", "GetNodeGroups", len(groups), elapsed)
	for i, g := range groups {
		if config.verbose {
			fmt.Printf("\t%d group [%v]\n", i, g.Name)
		}
	}
	totalElapsed += elapsed
}

func getRoles() {
	start := time.Now()
	roles, err := Pcc.GetRoles()
	elapsed := time.Since(start)
	if err != nil {
		fmt.Printf("Error GetRoles: %v\n", err)
		return
	}
	if config.verbose {
		fmt.Println("")
	}
	fmt.Printf("%-18s %3d = %v\n", "GetRoles", len(roles), elapsed)
	for i, r := range roles {
		if config.verbose {
			fmt.Printf("\t%d role [%v]\n", i, r.Name)
		}
	}
	totalElapsed += elapsed
}

func getTemplates() {
	start := time.Now()
	templates, err := Pcc.GetRoleTemplates()
	elapsed := time.Since(start)
	if err != nil {
		fmt.Printf("Error GetRoleTemplates: %v\n", err)
		return
	}
	if config.verbose {
		fmt.Println("")
	}
	fmt.Printf("%-18s %3d = %v\n", "GetTemplates", len(templates), elapsed)
	for i, t := range templates {
		if config.verbose {
			fmt.Printf("\t%d template [%+v]\n", i, t.Name)
		}
	}
	totalElapsed += elapsed
}

func getNodes() {
	start := time.Now()
	nodes, err := Pcc.GetNodes()
	elapsed := time.Since(start)
	if err != nil {
		fmt.Printf("Error GetNodes: %v\n", err)
		return
	}
	if config.verbose {
		fmt.Println("")
	}
	fmt.Printf("%-18s %3d = %v\n", "GetNodes", len(*nodes), elapsed)
	for i, n := range *nodes {
		if config.verbose {
			fmt.Printf("\t%d node %d %v [%v]\n",
				i, n.Id, n.Name, n.Host)

		}
	}
	totalElapsed += elapsed
}

func getNodesOptions() {
	start := time.Now()
	nodes, err := Pcc.GetNodesOptions(0, 50, "name,clusterName",
		"asc,asc", "")
	elapsed := time.Since(start)
	if err != nil {
		fmt.Printf("Error GetNodes: %v\n", err)
		return
	}
	if config.verbose {
		fmt.Println("")
	}
	fmt.Printf("%-18s %3d = %v\n", "GetNodesOptions", len(*nodes), elapsed)
	for i, n := range *nodes {
		if config.verbose {
			fmt.Printf("\t%d node %d %-20s %-15s [%v]\n",
				i, n.Id, n.Name, n.Host, n.ProvisionStatus)
		}
	}
	totalElapsed += elapsed
}

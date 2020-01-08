package main

import (
	"fmt"
	"testing"

	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
)

func addSite(t *testing.T) {
	t.Run("addSite", addSites)
	t.Run("delSite", delSites)
	t.Run("addAssignSite", addAssignSite)
}

func addSites(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	fmt.Println("Add sites")
	for _, x := range []struct {
		name        string
		description string
	}{
		{"SJC", "San Jose"},
		{"ORD", "Chicago"},
		{"DEN", "Denver"},
		{"OSL", "Oslo"},
	} {
		fmt.Printf("add site %v\n", x.name)
		err := Pcc.AddSite(pcc.Site{Name: x.name,
			Description: x.description})
		if err != nil {
			assert.Fatalf("%v\n", err)
		}
	}

	if site, err := Pcc.FindSite("OSL"); err != nil {
		assert.Fatalf("%v\n", err)
	} else {
		fmt.Printf("found site %v\n", site.Name)
	}
}

func delSites(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	var (
		sites []pcc.Site
		err   error
	)
	fmt.Println("Get sites")
	sites, err = Pcc.GetSites()
	if err != nil {
		assert.Fatalf("%v\n", err)
	}
	for _, s := range sites {
		fmt.Printf("found site %v\n", s.Name)
		err = Pcc.DelSite(s)
		if err != nil {
			assert.Fatalf("%v\n", err)
		}
	}
}

func addAssignSite(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	addReq := pcc.Site{
		Name:        "SJC",
		Description: "San Jose",
	}
	fmt.Printf("add site %v\n", addReq.Name)
	err := Pcc.AddSite(addReq)
	if err != nil {
		assert.Fatalf("%v\n", err)
	}

}

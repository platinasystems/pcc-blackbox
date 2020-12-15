package main

import (
	"fmt"
	"time"
)

func getEventAction() {
	dockerStats.ChangePhase("getEvent")
	getEvent()
}

func getEvent() {
	var (
		err    error
		page   int = 0
		limit  int = 50
		search string
	)

	if config.search != "" {
		search = config.search
	}
	if config.page != 0 {
		page = config.page
	}
	if config.limit != 50 {
		limit = config.limit
	}
	start := time.Now()
	events, err := Pcc.GetEventsOptions(page, limit, search)
	elapsed := time.Since(start)
	if err != nil {
		fmt.Printf("Error GetEvents: %v\n", err)
		return
	}
	fmt.Printf("%v events found\n", len(events))
	for i, e := range events {
		uN := e.CreatedAt / 1000
		cA := time.Unix(int64(uN), 0)
		typeLevel := e.Type + ":" + e.Level
		if config.verbose {
			fmt.Printf("%-3d %v\n", i+1, e)
		} else {
			fmt.Printf("%-3d %-15s %-20s %-50s %v\n",
				i+1, e.TargetName, typeLevel, e.Message,
				cA.Format("01/02 15:04:00 MST"))
		}
	}
	fmt.Printf("\nelapsed %v\n", elapsed)
}

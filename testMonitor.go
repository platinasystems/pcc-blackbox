package main

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

var optionalTopics = []string{"ceph-metrics", "varnishStats", "ksm", "flowStats", "nodeDetails-2", "podDetails-2", "svcDetails-2", "appDetails"}

// get topics
func testGetTopic(t *testing.T) {
	if topics, err := Pcc.GetTopics(); err == nil {
		if len(topics) == 0 {
			t.Fatal("unable to fetch topic names")
		} else {
			fmt.Println(fmt.Sprintf("get topic %+v", topics))
		}
	} else {
		t.Fatal(err)
	}
}

// get topics schema
func testGetTopicSchema(t *testing.T) {
	if topics, err := Pcc.GetTopics(); err == nil {
	l1:
		for _, topic := range topics {
			if schema, err := Pcc.GetSchema(topic); err == nil {
				fmt.Println(fmt.Sprintf("there are %d schemas for topic %s", len(schema), topic))
			} else {
				for _, optionalTopic := range optionalTopics {
					if optionalTopic == topic {
						fmt.Println(fmt.Sprintf("topic %s not created yet", topic))
						continue l1
					}
				}
				t.Fatal(err)
			}
		}
	} else {
		t.Fatal(err)
	}
}

// get a sample and check the content
func testMonitorSample(t *testing.T) {
	if nodes, err := Pcc.GetNodes(); err == nil {
		node := (*nodes)[0]
		if data, err := Pcc.GetLiveSample("cpu", node.Id); err == nil {
			fmt.Println(fmt.Sprintf("read cpu sample for node %d %+v", node.Id, data))
			fmt.Println(data)
			if nodeId, err := strconv.ParseUint(fmt.Sprintf("%v", (*data)["nodeId"]), 10, 64); err == nil {
				if nodeId != node.Id {
					t.Fatal("get the wrong sample")
				}
			} else {
				t.Fatal(err)
			}
		} else {
			t.Fatal(err)
		}
	} else {
		t.Fatal(err)
	}
}

// get history
func testMonitorHistory(t *testing.T) {
	to := time.Now().Unix()
	from := to - (1000 * 60 * 100)

	if data, err := Pcc.GetHistory("cpu", from, to); err == nil {
		fmt.Println(fmt.Sprintf("read cpu history %+v", data)) // nothing to do
	} else {
		t.Fatal(err)
	}
}

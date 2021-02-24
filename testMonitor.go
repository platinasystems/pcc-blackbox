package main

import (
	"fmt"
	log "github.com/platinasystems/go-common/logs"
	"github.com/platinasystems/pcc-blackbox/models"
	"strconv"
	"testing"
	"time"
)

var optionalTopics = []string{"ceph-metrics", "varnishStats", "ksm", "flowStats", "nodeDetails-2", "podDetails-2", "svcDetails-2", "appDetails"}

// get topics
func testGetTopic(t *testing.T) {

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now(), "testGetTopic")

	if topics, err := Pcc.GetTopics(); err == nil {
		if len(topics) == 0 {
			msg := "unable to fetch topic names"
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			t.FailNow()
		} else {
			log.AuctaLogger.Info(fmt.Sprintf("get topic %+v", topics))
		}
	} else {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}
}

// get topics schema
func testGetTopicSchema(t *testing.T) {

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now(), "testGetTopicSchema")

	if topics, err := Pcc.GetTopics(); err == nil {
	l1:
		for _, topic := range topics {
			if schema, err := Pcc.GetSchema(topic); err == nil {
				log.AuctaLogger.Info(fmt.Sprintf("there are %d schemas for topic %s", len(schema), topic))
			} else {
				for _, optionalTopic := range optionalTopics {
					if optionalTopic == topic {
						log.AuctaLogger.Infof(fmt.Sprintf("topic %s not created yet", topic))
						continue l1
					}
				}
				msg := fmt.Sprintf("%v", err)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				t.FailNow()
			}
		}
	} else {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}
}

// get a sample and check the content
func testMonitorSample(t *testing.T) {

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now(), "testMonitorSample")

	if nodes, err := Pcc.GetNodes(); err == nil {
		node := (*nodes)[0]
		if data, err := Pcc.GetLiveSample("cpu", node.Id); err == nil {
			log.AuctaLogger.Info(fmt.Sprintf("read cpu sample for node %d %+v", node.Id, data))
			log.AuctaLogger.Info(data)
			if nodeId, err := strconv.ParseUint(fmt.Sprintf("%v", (*data)["nodeId"]), 10, 64); err == nil {
				if nodeId != node.Id {
					msg := "get the wrong sample"
					res.SetTestFailure(msg)
					log.AuctaLogger.Error(msg)
					t.FailNow()
				}
			} else {
				msg := fmt.Sprintf("%v", err)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				t.FailNow()
			}
		} else {
			msg := fmt.Sprintf("%v", err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			t.FailNow()
		}
	} else {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}
}

// get history
func testMonitorHistory(t *testing.T) {

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now(), "testMonitorHistory")

	to := time.Now().Unix()
	from := to - (1000 * 60 * 100)

	if data, err := Pcc.GetHistory("cpu", from, to); err == nil {
		log.AuctaLogger.Info(fmt.Sprintf("read cpu history %+v", data)) // nothing to do
	} else {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}
}

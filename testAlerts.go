package main

import (
	"errors"
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"github.com/mitchellh/mapstructure"
	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	m "github.com/platinasystems/pcc-blackbox/models"
	"github.com/platinasystems/test"
	"github.com/platinasystems/tiles/pccserver/models"
	"io"
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

const (
	HighUsageAlertName     = "ceph pools high usage"
	VeryHighUsageAlertName = "ceph pools very high usage"
)

var oldMail string
var poolName string

func testPoolUsageAlert(t *testing.T) {

	t.Run("changeRootMail", changeRootMail)
	t.Run("createPool", createPool)
	t.Run("verifyPool", verifyPool)
	t.Run("testEmptyPoolPrometheus", testEmptyPoolPrometheus)
	t.Run("addFileOver80", addFileOver80)
	t.Run("testPoolOver80Prometheus", testPoolOver80Prometheus)
	t.Run("testMonitorRuleFiring80", testMonitorRuleFiring80)
	t.Run("testNotificationRuleFiring80", testNotificationRuleFiring80)
	t.Run("testMail", testMail)
	t.Run("resetRootMail", resetRootMail)
}

func changeRootMail(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	user, err := Pcc.GetUser(1)
	checkError(t, res, err)

	oldMail = user.Profile.Email
	user.Email = "aucta.tenant@gmail.com"
	user.Role = nil

	err = Pcc.UpdateUser(user)
	checkError(t, res, err)
}

func testMail(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	c, err := client.DialTLS("imap.gmail.com:993", nil)
	checkError(t, res, err)
	defer c.Logout()

	err = c.Login("aucta.tenant@gmail.com", "plat1n@!")
	checkError(t, res, err)

	// Select INBOX
	mbox, err := c.Select("inbox", false)
	checkError(t, res, err)

	// Get the last 4 messages
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > 3 {
		// We're using unsigned integers here, only subtract if the result is > 0
		from = mbox.Messages - 3
	}
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	var section imap.BodySectionName

	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{section.FetchItem(), imap.FetchEnvelope}, messages)
	}()

	for msg := range messages {
		if msg.Envelope.Subject != "Monitoring Alert" {
			continue
		}
		r := msg.GetBody(&section)
		mr, err := mail.CreateReader(r)
		checkError(t, res, err)
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				log.AuctaLogger.Error(err)
			}

			switch h := p.Header.(type) {
			case *mail.InlineHeader:
				// This is the message's text (can be plain-text or HTML)
				b, _ := ioutil.ReadAll(p.Body)
				split := strings.Split(string(b), "<dt>Rule:</dt>")
				for _, s := range split {
					if strings.Contains(s, poolName) && strings.Contains(s, "firing") {
						return
					}
				}

			case *mail.AttachmentHeader:
				// This is an attachment
				filename, _ := h.Filename()
				log.AuctaLogger.Infof("Got attachment: %v", filename)
			}
		}
	}

	err = errors.New("Could not find the notification email")
	checkError(t, res, err)
}

func resetRootMail(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	user, err := Pcc.GetUser(1)
	checkError(t, res, err)

	user.Email = oldMail
	user.Role = nil

	err = Pcc.UpdateUser(user)
	checkError(t, res, err)
}

func createPool(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	poolName = fmt.Sprintf("pool%d", randomGenerator.Intn(999999999))
	var (
		err error
	)

	cluster, err = Pcc.GetCephCluster(Env.AlertsConfiguration.ClusterName)
	checkError(t, res, err)

	poolRequest := pcc.CreateCephPoolRequest{
		CephClusterId: cluster.Id,
		Name:          poolName,
		QuotaUnit:     "MiB",
		Quota:         "100",
		PoolType:      models.CEPH_POOL_PROFILE_TYPE_REPLICATED.String(),
		Size:          3,
	}

	_, err = Pcc.GetCephPool(poolName, cluster.Id)

	if err != nil {
		_, err = Pcc.CreateCephPool(poolRequest)
		checkError(t, res, err)
	} else {
		err = errors.New("Pool already exists")
		checkError(t, res, err)
	}
}

func verifyPool(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	cephCluster, err := Pcc.GetCephCluster(Env.AlertsConfiguration.ClusterName)
	checkError(t, res, err)

	timeout := time.After(15 * time.Minute)
	tick := time.Tick(30 * time.Second)
	for true {
		select {
		case <-timeout:
			msg := "Timed out waiting for pool response"
			checkError(t, res, errors.New(msg))
		case <-tick:
			pool, err := Pcc.GetCephPool(poolName, cephCluster.Id)
			if err != nil {
				msg := fmt.Sprintf("Failed to get deploy status "+
					"%v", err)
				checkError(t, res, errors.New(msg))
			}
			switch pool.DeployStatus {
			case pcc.RGW_DEPLOY_STATUS_PROGRESS:
				log.AuctaLogger.Info("pool installation in progress...")
			case pcc.RGW_DEPLOY_STATUS_COMPLETED:
				log.AuctaLogger.Info("pool installation completed")
				return
			case pcc.RGW_DEPLOY_STATUS_FAILED:
				msg := "pool installation failed"
				checkError(t, res, errors.New(msg))
			default:
				msg := fmt.Sprintf("Unexpected status - %v",
					pool.DeployStatus)
				checkError(t, res, errors.New(msg))
			}
		}
	}
}

func testEmptyPoolPrometheus(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	time.Sleep(1 * time.Minute)

	usage := checkPoolUsagePrometheus(t, res)
	log.AuctaLogger.Infof("usage: %v", usage)
	var err error
	if usage != 0 {
		err = errors.New("Pool should be empty")
	}
	checkError(t, res, err)
}

func checkPoolUsagePrometheus(t *testing.T, res *m.TestResult) int {
	query := fmt.Sprintf("pools:%s:quotaUsage", poolName)

	result, err := Pcc.InstantQuery(query)
	checkError(t, res, err)

	log.AuctaLogger.Infof("result: %v", result.Value)
	return int(result.Value)
}

func addFileOver80(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	cephCluster, err := Pcc.GetCephCluster(Env.AlertsConfiguration.ClusterName)
	checkError(t, res, err)
	targetNodeID := cephCluster.Nodes[0].NodeId

	poolNode, err := Pcc.GetNode(targetNodeID)
	checkError(t, res, err)
	targetNodeHost := poolNode.Host
	log.AuctaLogger.Info(targetNodeHost)

	_, _, err = Pcc.SSHHandler().Run(targetNodeHost, "fallocate -l 85MiB 85MiB_file")
	checkError(t, res, err)

	_, _, err = Pcc.SSHHandler().Run(targetNodeHost, fmt.Sprintf("sudo rados -p %s put 85MiB_file 85MiB_file", poolName))
	checkError(t, res, err)

	_, _, err = Pcc.SSHHandler().Run(targetNodeHost, "rm 85MiB_file")
	checkError(t, res, err)
}

func testPoolOver80Prometheus(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	log.AuctaLogger.Info("Waiting for the alert to be propagated (2m)...")
	time.Sleep(2 * time.Minute)

	usage := checkPoolUsagePrometheus(t, res)

	var err error
	if usage < 80 {
		err = errors.New("Pool usage should be over 80%")
	}
	checkError(t, res, err)
}

func testMonitorRuleFiring80(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	rules, err := Pcc.GetRules()
	checkError(t, res, err)

	log.AuctaLogger.Info(rules)

	for _, r := range *rules {
		if r.Name == HighUsageAlertName {
			if len(r.Alerts) == 0 {
				err = errors.New("The alert list shouldn't be empty")
				checkError(t, res, err)
			}
			return
		}
	}
}

func testNotificationRuleFiring80(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	notifications, err := Pcc.GetNotifications()
	checkError(t, res, err)

	for _, n := range notifications {
		var meta pcc.AlertMetadata
		mapstructure.Decode(n.Metadata, &meta)
		if strings.Contains(meta.Description, poolName) && meta.Status == "firing" {
			return
		}
	}

	err = errors.New("Could not find the alert firing notification")
	checkError(t, res, err)
}

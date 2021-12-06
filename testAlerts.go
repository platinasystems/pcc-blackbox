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
	"github.com/platinasystems/pcc-models/notification"
	"github.com/platinasystems/test"
	"github.com/platinasystems/tiles/pccserver/models"
	//"github.com/platinasystems/pcc-models/notification"
	monitorModels "github.com/platinasystems/platina-monitor/models"
	"io"
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

const (
	HighUsageAlertName     = "ceph pools high usage"
	VeryHighUsageAlertName = "ceph pools very high usage"
	OSDDownOut             = "osds down/out"
)

var (
	oldMail            string
	poolName           string
	pool1, pool2       string
	osdServer, osdHost string
	osdID              int
	savedNotifications map[string][]monitorModels.RuleNotification
)

func test2PoolUsageAlert(t *testing.T) {
	t.Run("testChangeMailAlerts", testChangeMailAlerts)

	t.Run("createPool", createPool)
	t.Run("verifyPool", verifyPool)
	pool1 = poolName
	t.Run("createPool", createPool)
	t.Run("verifyPool", verifyPool)
	pool2 = poolName

	waitMinutes(1 * time.Minute)

	usePool1()
	t.Run("testEmptyPoolPrometheus", testEmptyPoolPrometheus)
	t.Run("addFile85MiB", addFile85MiB)
	usePool2()
	t.Run("testEmptyPoolPrometheus", testEmptyPoolPrometheus)
	t.Run("addFile85MiB", addFile85MiB)

	waitMinutes(3 * time.Minute)

	usePool1()
	t.Run("testPoolOver80Prometheus", testPoolOver80Prometheus)
	t.Run("testMonitorRuleFiring80", testMonitorRuleFiring80)
	t.Run("testNotificationRuleFiring80", testNotificationRuleFiring80)
	t.Run("testMailFiring80", testMailFiring80)
	t.Run("addFile10MiB", addFile10MiB)
	usePool2()
	t.Run("testPoolOver80Prometheus", testPoolOver80Prometheus)
	t.Run("testMonitorRuleFiring80", testMonitorRuleFiring80)
	t.Run("testNotificationRuleFiring80", testNotificationRuleFiring80)
	t.Run("testMailFiring80", testMailFiring80)
	t.Run("addFile10MiB", addFile10MiB)

	waitMinutes(3 * time.Minute)

	usePool1()
	t.Run("testPoolOver90Prometheus", testPoolOver90Prometheus)
	t.Run("testMonitorRuleFiring90", testMonitorRuleFiring90)
	t.Run("testNotificationRuleFiring90", testNotificationRuleFiring90)
	t.Run("testMailFiring90", testMailFiring90)
	t.Run("delAllFiles", delAllFiles)

	usePool2()
	t.Run("testPoolOver90Prometheus", testPoolOver90Prometheus)
	t.Run("testMonitorRuleFiring90", testMonitorRuleFiring90)
	t.Run("testNotificationRuleFiring90", testNotificationRuleFiring90)
	t.Run("testMailFiring90", testMailFiring90)
	t.Run("delAllFiles", delAllFiles)

	waitMinutes(3 * time.Minute)

	usePool1()
	t.Run("testEmptyPoolPrometheus", testEmptyPoolPrometheus)
	t.Run("testMonitorRuleResolved80", testMonitorRuleResolved80)
	t.Run("testMonitorRuleResolved90", testMonitorRuleResolved90)
	t.Run("testNotificationRuleResolved80", testNotificationRuleResolved80)
	t.Run("testNotificationRuleResolved90", testNotificationRuleResolved90)
	t.Run("testMailResolved80", testMailResolved80)
	t.Run("testMailResolved90", testMailResolved90)
	t.Run("removePool", removePool)

	usePool2()
	t.Run("testEmptyPoolPrometheus", testEmptyPoolPrometheus)
	t.Run("testMonitorRuleResolved80", testMonitorRuleResolved80)
	t.Run("testMonitorRuleResolved90", testMonitorRuleResolved90)
	t.Run("testNotificationRuleResolved80", testNotificationRuleResolved80)
	t.Run("testNotificationRuleResolved90", testNotificationRuleResolved90)
	t.Run("testMailResolved80", testMailResolved80)
	t.Run("testMailResolved90", testMailResolved90)
	t.Run("removePool", removePool)

	t.Run("restoreMailAlerts", restoreMailAlerts)
}

func testOSDDownAlert(t *testing.T) {
	t.Run("testChangeMailAlerts", testChangeMailAlerts)

	t.Run("testOSDDown", testOSDDown)

	waitMinutes(10 * time.Minute)

	t.Run("testOSDDownPrometheus", testOSDDownPrometheus)
	t.Run("testMonitorRuleFiringOSD", testMonitorRuleFiringOSD)
	t.Run("testNotificationRuleFiringOSD", testNotificationRuleFiringOSD)
	t.Run("testMailFiringOSD", testMailFiringOSD)

	t.Run("testOSDUp", testOSDUp)

	waitMinutes(3 * time.Minute)

	t.Run("testOSDUpPrometheus", testOSDUpPrometheus)
	t.Run("testMonitorRuleResolvedOSD", testMonitorRuleResolvedOSD)
	t.Run("testNotificationRuleResolvedOSD", testNotificationRuleResolvedOSD)
	t.Run("testMailResolvedOSD", testMailResolvedOSD)

	t.Run("restoreMailAlerts", restoreMailAlerts)
}

func testChangeMailAlerts(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	savedNotifications = make(map[string][]monitorModels.RuleNotification)

	for _, name := range []string{HighUsageAlertName, VeryHighUsageAlertName, OSDDownOut} {
		rule, err := Pcc.GetRuleByName(name)
		checkError(t, res, err)

		savedNotifications[name] = rule.Notifications

		notification := monitorModels.RuleNotification{
			NotificationService: notification.NotificationService{
				Service: "email",
				Inputs: []notification.NotificationInput{
					{
						Name:  "to",
						Value: Env.AlertsConfiguration.MailUsername,
					},
				},
			},
		}

		rule.Notifications = []monitorModels.RuleNotification{notification}

		_, err = Pcc.UpdateRule(rule, rule.Id)
		checkError(t, res, err)
	}
}

func restoreMailAlerts(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	for _, name := range []string{HighUsageAlertName, VeryHighUsageAlertName, OSDDownOut} {
		rule, err := Pcc.GetRuleByName(name)
		checkError(t, res, err)

		notifications := savedNotifications[name]

		for n := range notifications {
			notifications[n].Id = 0
		}
		rule.Notifications = notifications

		_, err = Pcc.UpdateRule(rule, rule.Id)
		checkError(t, res, err)
	}
}

func usePool1() {
	poolName = pool1
}

func usePool2() {
	poolName = pool2
}

func waitMinutes(minutes time.Duration) {
	log.AuctaLogger.Info(fmt.Sprintf("Waiting %d m", minutes/time.Minute))
	time.Sleep(minutes)
}

func changeRootMail(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	user, err := Pcc.GetUser(1)
	checkError(t, res, err)

	oldMail = user.Profile.Email
	user.Email = Env.AlertsConfiguration.MailUsername
	user.Role = nil

	err = Pcc.UpdateUser(user)
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

func removePool(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	cephCluster, err := Pcc.GetCephCluster(Env.AlertsConfiguration.ClusterName)
	checkError(t, res, err)

	pool, err := Pcc.GetCephPool(poolName, cephCluster.Id)
	err = Pcc.DeleteCephPool(pool.Id)
	checkError(t, res, err)
}

func testOSDDown(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	cluster, err := Pcc.GetCephCluster(Env.AlertsConfiguration.ClusterName)
	checkError(t, res, err)
	osds, err := Pcc.GetOSDsStateByClusterID(cluster.Id)
	checkError(t, res, err)
	targetOSD := osds[0]

	osdNode, err := Pcc.GetNodeByName(targetOSD.Server)
	checkError(t, res, err)

	osdHost = osdNode.Host
	osdServer = targetOSD.Server
	osdID = targetOSD.Osd

	_, _, err = Pcc.SSHHandler().Run(osdHost, fmt.Sprintf("sudo systemctl stop ceph-osd@%d.service", osdID))
	checkError(t, res, err)
}

func testOSDUp(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	_, _, err := Pcc.SSHHandler().Run(osdHost, fmt.Sprintf("sudo systemctl start ceph-osd@%d.service", osdID))
	checkError(t, res, err)
}

func addFiles(t *testing.T, res *m.TestResult, poolName string, fileSize ...string) {
	cephCluster, err := Pcc.GetCephCluster(Env.AlertsConfiguration.ClusterName)
	checkError(t, res, err)
	targetNodeID := cephCluster.Nodes[0].NodeId

	poolNode, err := Pcc.GetNode(targetNodeID)
	checkError(t, res, err)
	targetNodeHost := poolNode.Host
	log.AuctaLogger.Info(targetNodeHost)

	for _, f := range fileSize {
		_, _, err = Pcc.SSHHandler().Run(targetNodeHost, fmt.Sprintf("fallocate -l %s %s_file", f, f))
		checkError(t, res, err)

		_, _, err = Pcc.SSHHandler().Run(targetNodeHost, fmt.Sprintf("sudo rados -p %s put %s_file %s_file", poolName, f, f))
		checkError(t, res, err)

		_, _, err = Pcc.SSHHandler().Run(targetNodeHost, fmt.Sprintf("rm %s_file", f))
		checkError(t, res, err)
	}
}

func delFiles(t *testing.T, res *m.TestResult, poolName string, fileSize ...string) {
	cephCluster, err := Pcc.GetCephCluster(Env.AlertsConfiguration.ClusterName)
	checkError(t, res, err)
	targetNodeID := cephCluster.Nodes[0].NodeId

	poolNode, err := Pcc.GetNode(targetNodeID)
	checkError(t, res, err)
	targetNodeHost := poolNode.Host
	log.AuctaLogger.Info(targetNodeHost)

	for _, f := range fileSize {
		_, _, err = Pcc.SSHHandler().Run(targetNodeHost, fmt.Sprintf("sudo rados -p %s rm %s_file", poolName, f))
		checkError(t, res, err)
	}
}

func addFile85MiB(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	addFiles(t, res, poolName, "85MiB")
}

func addFile10MiB(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	addFiles(t, res, poolName, "10MiB")
}

func delAllFiles(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	delFiles(t, res, poolName, "85MiB", "10MiB")
}

func checkPoolUsagePrometheus(t *testing.T, res *m.TestResult) int {
	query := fmt.Sprintf("pools:%s:quotaUsage", poolName)

	result, err := Pcc.InstantQuery(query)
	checkError(t, res, err)

	log.AuctaLogger.Infof("result: %v", result.Value)
	return int(result.Value)
}

func testEmptyPoolPrometheus(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	usage := checkPoolUsagePrometheus(t, res)
	log.AuctaLogger.Infof("usage: %v", usage)
	var err error
	if usage != 0 {
		err = errors.New("Pool should be empty")
	}
	checkError(t, res, err)
}

func testPoolOver80Prometheus(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	log.AuctaLogger.Info("Waiting for the alert to be propagated (2m)...")

	usage := checkPoolUsagePrometheus(t, res)

	var err error
	if usage < 80 {
		err = errors.New("Pool usage should be over 80%")
	}
	checkError(t, res, err)
}

func testPoolOver90Prometheus(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	usage := checkPoolUsagePrometheus(t, res)

	var err error
	if usage < 90 {
		err = errors.New("Pool usage should be over 90%")
	}
	checkError(t, res, err)
}

func checkDownOSDsNumPrometheus(t *testing.T, res *m.TestResult) int {
	query := "numOsdsNotUp"

	result, err := Pcc.InstantQuery(query)
	checkError(t, res, err)

	log.AuctaLogger.Infof("result: %v", result.Value)
	return int(result.Value)
}

func testOSDDownPrometheus(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	downOSDs := checkDownOSDsNumPrometheus(t, res)

	var err error
	if downOSDs == 0 {
		err = errors.New("There should be at least an OSD down/out")
	}
	checkError(t, res, err)
}

func testOSDUpPrometheus(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	downOSDs := checkDownOSDsNumPrometheus(t, res)

	var err error
	if downOSDs > 0 {
		err = errors.New("All OSDs should be up")
	}
	checkError(t, res, err)
}

func testMonitorRuleFiring(t *testing.T, res *m.TestResult, ruleName string, objName string) {
	rules, err := Pcc.GetRules()
	checkError(t, res, err)

	for _, r := range *rules {
		if r.Name == ruleName {
			for _, a := range r.Alerts {
				if a.ObjName == objName || objName == "" {
					return
				}
			}
		}
	}

	err = errors.New("Could not find the alert")
	checkError(t, res, err)
}

func testMonitorRuleFiring80(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	testMonitorRuleFiring(t, res, HighUsageAlertName, poolName)
}

func testMonitorRuleFiring90(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	testMonitorRuleFiring(t, res, VeryHighUsageAlertName, poolName)
}

func testMonitorRuleFiringOSD(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	testMonitorRuleFiring(t, res, OSDDownOut, osdServer)
}

func testMonitorRuleResolved(t *testing.T, res *m.TestResult, ruleName string, objName string) {
	rules, err := Pcc.GetRules()
	checkError(t, res, err)

	for _, r := range *rules {
		if r.Name == ruleName {
			for _, a := range r.Alerts {
				if a.ObjName == objName || objName == "" {
					err = errors.New("Alert still active in monitoring")
					checkError(t, res, err)
					return
				}
			}
		}
	}
	return
}

func testMonitorRuleResolved80(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	testMonitorRuleResolved(t, res, HighUsageAlertName, poolName)
}

func testMonitorRuleResolved90(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	testMonitorRuleResolved(t, res, VeryHighUsageAlertName, poolName)
}

func testMonitorRuleResolvedOSD(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	testMonitorRuleResolved(t, res, OSDDownOut, osdServer)
}

func testNotifications(t *testing.T, res *m.TestResult, status string, alert string, objName string) {
	notifications, err := Pcc.GetNotifications()
	checkError(t, res, err)

	for _, n := range notifications {
		var meta pcc.AlertMetadata
		mapstructure.Decode(n.Metadata, &meta)
		if strings.Contains(meta.Description, alert) && strings.Contains(meta.Description, objName) && meta.Status == status {
			return
		}
	}

	err = errors.New("Could not find the alert firing notification")
	checkError(t, res, err)
}

func testNotificationRuleFiring80(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	testNotifications(t, res, "firing", "80%", poolName)
}

func testNotificationRuleFiring90(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	testNotifications(t, res, "firing", "90%", poolName)
}

func testNotificationRuleFiringOSD(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	testNotifications(t, res, "firing", "OSDs", osdServer)
}

func testNotificationRuleResolved80(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	testNotifications(t, res, "resolved", "80%", poolName)
}

func testNotificationRuleResolved90(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	testNotifications(t, res, "resolved", "90%", poolName)
}

func testNotificationRuleResolvedOSD(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	testNotifications(t, res, "resolved", "OSDs", osdServer)
}

func testMailKeywords(t *testing.T, res *m.TestResult, keywords ...string) {
	c, err := client.DialTLS(Env.AlertsConfiguration.MailIMAP, nil)
	checkError(t, res, err)
	defer c.Logout()

	err = c.Login(Env.AlertsConfiguration.MailUsername, Env.AlertsConfiguration.MailPassword)
	checkError(t, res, err)

	// Select INBOX
	mbox, err := c.Select("inbox", false)
	checkError(t, res, err)

	// Get the last 4 messages
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > 10 {
		// We're using unsigned integers here, only subtract if the result is > 0
		from = mbox.Messages - 10
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
				body := string(b)
				ok := true
				for _, k := range keywords {
					ok = ok && strings.Contains(body, k)
				}
				if ok {
					return
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

func testMailFiring80(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	testMailKeywords(t, res, "80%", "firing", poolName)
}

func testMailFiring90(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	testMailKeywords(t, res, "90%", "firing", poolName)
}

func testMailFiringOSD(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	testMailKeywords(t, res, "OSDs", "firing", osdServer)
}

func testMailResolved80(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	testMailKeywords(t, res, "80%", "resolved", poolName)
}

func testMailResolved90(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	testMailKeywords(t, res, "90%", "resolved", poolName)
}

func testMailResolvedOSD(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	testMailKeywords(t, res, "OSDs", "resolved", osdServer)
}

package pcc

import (
	"time"
	"fmt"
	"github.com/platinasystems/tiles/pccserver/models"
	"encoding/json"
	"strings"
)

const (
	FREQUENCY            = 10

	CEPH_CLUSTER_INSTALL_EVENT 	= "cephClusterInstall"
	CEPH_CLUSTER_UNINSTALL_EVENT 	= "cephClusterUninstall"
	CEPH_POOL_CREATE_EVENT 		= "cephPoolCreate"
	CEPH_POOL_DELETE_EVENT 		= "cephPoolDelete"
	CEPH_FS_CREATE_EVENT 		= "cephFSCreate"
	CEPH_FS_DELETE_EVENT 		= "cephFSDelete"
)

var EventVerificationMap = map[string]func(string) []verifier{
	CEPH_CLUSTER_INSTALL_EVENT: func(name string) []verifier{
		return []verifier{
			{
				timeout: CEPH_3_NODE_INSTALLATION_TIMEOUT,
				notification: notification {
					CEPH_INSTALLATION_SUCCESS_NOTIFICATION : true,
					fmt.Sprintf(CEPH_INSTALLATION_FAILED_NOTIFICATION_1, GetPccClient().GetCephClusterName()): true,
					CEPH_INSTALLATION_FAILED_NOTIFICATION_2: true,
					CEPH_INSTALLATION_FAILED_NOTIFICATION_3: true,
					fmt.Sprintf(CEPH_INSTALLATION_FAILED_NOTIFICATION_4, GetPccClient().GetCephClusterName()): true,
					CEPH_INSTALLATION_INTERMEDIATE_NOTIFICATION_1: false,
					CEPH_INSTALLATION_INTERMEDIATE_NOTIFICATION_2: false,
					CEPH_INSTALLATION_INTERMEDIATE_NOTIFICATION_3: false,
					CEPH_INSTALLATION_INTERMEDIATE_NOTIFICATION_4: false,
					fmt.Sprintf(CEPH_INSTALLATION_INTERMEDIATE_NOTIFICATION_5, GetPccClient().GetCephClusterName()): false,
					CEPH_INSTALLATION_INTERMEDIATE_NOTIFICATION_6: false,
					CEPH_INSTALLATION_INTERMEDIATE_NOTIFICATION_7: false,
					CEPH_INSTALLATION_INTERMEDIATE_NOTIFICATION_8: false,
				},
				breakLoop: make(chan bool),
			},
		}
	},
	CEPH_CLUSTER_UNINSTALL_EVENT: func(name string) []verifier{
		return []verifier{
			{
				timeout: CEPH_3_NODE_UNINSTALLATION_TIMEOUT,
				notification: notification{
					CEPH_UNINSTALLATION_SUCCESS_NOTIFICATION: true,
					fmt.Sprintf(CEPH_UNINSTALLATION_FAILED_NOTIFICATION_1, GetPccClient().GetCephClusterName()): true,
					fmt.Sprintf(CEPH_UNINSTALLATION_FAILED_NOTIFICATION_2, GetPccClient().GetCephClusterName()): true,
					CEPH_UNINSTALLATION_INTERMEDIATE_NOTIFICATION_1: false,
					CEPH_UNINSTALLATION_INTERMEDIATE_NOTIFICATION_2: false,
					CEPH_UNINSTALLATION_INTERMEDIATE_NOTIFICATION_3: false,
					CEPH_UNINSTALLATION_INTERMEDIATE_NOTIFICATION_4: false,
				},
				breakLoop: make(chan bool),
			},
		}
	},
	CEPH_POOL_CREATE_EVENT: func(name string) []verifier{
		return []verifier {
			{
				timeout: CEPH_POOL_CREATION_TIMEOUT,
				notification: notification{
					fmt.Sprintf(CEPH_POOL_CREATION_SUCCESS_NOTIFICATION, name, GetPccClient().GetCephClusterName()): true,
					fmt.Sprintf(CEPH_POOL_CREATION_FAILED_NOTIFICATION, name, GetPccClient().GetCephClusterName()): true,
					fmt.Sprintf(CEPH_POOL_CREATION_INTERMEDIATE_NOTIFICATION_1, name, GetPccClient().GetCephClusterName()): false,
				},
				breakLoop: make(chan bool),
			},
		}
	},
	CEPH_POOL_DELETE_EVENT: func(name string) []verifier{
		return []verifier {
			{
				timeout: CEPH_POOL_DELETION_TIMEOUT,
				notification: notification{
					fmt.Sprintf(CEPH_POOL_DELETION_SUCCESS_NOTIFICATION, name): true,
					fmt.Sprintf(CEPH_POOL_DELETION_FAILED_NOTIFICATION, name): true,
					fmt.Sprintf(CEPH_POOL_DELETION_INTERMEDIATE_NOTIFICATION_1, name, GetPccClient().GetCephClusterName()): false,
				},
				breakLoop: make(chan bool),
			},
		}
	},
	CEPH_FS_CREATE_EVENT: func(name string) []verifier{
		return []verifier {
			{
				timeout: CEPH_FS_CREATION_TIMEOUT,
				notification: notification{
					fmt.Sprintf(CEPH_FS_CREATION_SUCCESS_NOTIFICATION, name, GetPccClient().GetCephClusterName()): true,
					fmt.Sprintf(CEPH_FS_CREATION_FAILED_NOTIFICATION_1, name, GetPccClient().GetCephClusterName()): true,
					fmt.Sprintf(CEPH_FS_CREATION_FAILED_NOTIFICATION_2, name, GetPccClient().GetCephClusterName()): true,
					fmt.Sprintf(CEPH_FS_CREATION_INTERMEDIATE_NOTIFICATION_1, name, GetPccClient().GetCephClusterName()): false,
				},
				breakLoop: make(chan bool),
			},
		}
	},
	CEPH_FS_DELETE_EVENT: func(name string) []verifier{
		return []verifier {
			{
				timeout: CEPH_FS_DELETION_TIMEOUT,
				notification: notification{
					fmt.Sprintf(CEPH_FS_DELETION_SUCCESS_NOTIFICATION, name): true,
					fmt.Sprintf(CEPH_FS_DELETION_FAILED_NOTIFICATION_1, name): true,
					fmt.Sprintf(CEPH_FS_DELETION_FAILED_NOTIFICATION_2, name): true,
					fmt.Sprintf(CEPH_FS_DELETION_INTERMEDIATE_NOTIFICATION_1, name): false,
				},
				breakLoop: make(chan bool),
			},
		}
	},
}

type Status struct {
	Msg string
	IsError bool
}
// For logging the notifications to stdout
// map[<notification_substring>]<do or dont terminate after notification is found>
type notification map[string]bool

type verifier struct {
	timeout time.Duration
	notification notification
	breakLoop chan bool
}

func Verify(startTime time.Time, action string, name string) (s Status){
	ver := EventVerificationMap[action](name)
	done := make(chan Status)
	for _, v := range ver {
		go syncCheckGenericInstallation(0, v.timeout, v.notification, startTime, done, v.breakLoop)
	}
	s = <- done
	go func() {
		for _, v := range ver {
			if v.breakLoop != nil {
				v.breakLoop <- true
			}
		}
	}()

	return
}

// Synchronize checking for installation
func syncCheckGenericInstallation(id uint64, appTimeout time.Duration, notifications2Check notification, from time.Time, found chan Status, breakLoop chan bool) {
	s := Status{}
	timeout := appTimeout * time.Second
	for time.Since(from) < timeout {
		select {
		case <- breakLoop:
			return
		default:
			var (
				events    []models.Notification
				err       error
			)
			events, err = getEvents()
			if err != nil {
				s.Msg = fmt.Sprintf("failed to getEvents ERROR: %v", err)
				s.IsError = true
				found <- s
			}else {
				eventsLoop:
				for i := 0; i < len(events); i++ {
					if events[i].CreatedAt < ConvertToMillis(from) {
						continue
					}
					for msg, terminate := range notifications2Check {
						eventMsg := events[i].Message
						if strings.Contains(eventMsg, msg) {
							if terminate {
								s.Msg = fmt.Sprintf("\"%v\" notification found in events", eventMsg)
								s.IsError = false
								found <- s
								break eventsLoop
							}else {
								fmt.Println("notification: ", eventMsg)
								delete(notifications2Check, msg)
							}
						}
					}
				}
			}
		}
		time.Sleep(FREQUENCY * time.Second)
	}

	s.Msg = fmt.Sprint("timeout exceeded")
	s.IsError = true
	found <- s
}

func getEvents() (events []models.Notification, err error) {
	var (
		resp HttpResp
	)
	resp, _, err = GetPccClient().pccGateway("GET", "pccserver/notifications/history?page=0&limit=50", nil)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(resp.Data, &events)
	if err != nil {
		return nil, err
	}
	return events, nil
}
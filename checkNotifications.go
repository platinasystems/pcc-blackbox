package main

import (
	"fmt"
	"github.com/platinasystems/tiles/pccserver/models"
	"strings"
	"time"
)

const (
	FREQUENCY                            = 10
	AGENT_TIMEOUT                        = 150
	COLLECTOR_TIMEOUT                    = 150
	LLDP_TIMEOUT                         = 300
	MAAS_INSTALL_TIMEOUT                 = 300
	K8S_INSTALL_TIMEOUT                  = 1800
	PORTUS_TIMEOUT                       = 400
	PXEBOOT_TIMEOUT                      = 400
	PORTUS_NOTIFICATION                  = "PORTUS"
	COLLECTOR_NOTIFICATION               = "COLLECTOR"
	AGENT_NOTIFICATION                   = "AGENT"
	LLDP_NOTIFICATION                    = "LLDPD"
	PXEBOOT_NODE_ADD_NOTIFICATION        = "new node added successfully"
	PXEBOOT_NODE_ADD_FAILED_NOTIFICATION = "add node at failed"
)

type status struct {
	msg     string
	isError bool
}

// Synchronize checking for installation
func syncCheckGenericInstallation(id uint64, appTimeout time.Duration, str2Check string, from time.Time, found chan status, breakLoop chan bool) {
	s := status{}
	timeout := appTimeout * time.Second
	for time.Since(from) < timeout {
		select {
		case <-breakLoop:
			return
		default:
			var (
				events []models.Notification
				err    error
			)
			events, err = getEvents()
			if err != nil {
				s.msg = fmt.Sprintf("failed to getEvents ERROR: %v", err)
				s.isError = true
				found <- s
			} else {
				for i := 0; i < len(events); i++ {
					if events[i].CreatedAt < ConvertToMillis(from) {
						continue
					}
					if strings.Contains(events[i].Message, str2Check) {
						s.msg = fmt.Sprintf("\"%v\" notification found in events", str2Check)
						s.isError = false
						found <- s
						break
					}
				}
			}
		}
		time.Sleep(FREQUENCY * time.Second)
	}

	s.msg = fmt.Sprint("timeout exceeded")
	s.isError = true
	found <- s
}

func getEvents() (events []models.Notification, err error) {
	err = Pcc.Get("pccserver/notifications/history", &events)
	return
}

func checkGenericInstallation(id uint64, appTimeout time.Duration, app string) (found bool, err error) {
	var (
		from    = time.Now()
		timeout = appTimeout * time.Second
	)
	str2Check := fmt.Sprintf("[%s] has been installed", app)
	return checkGenericEvent(timeout, id, str2Check, &from)
}

// Checks for notification and take care of the notification expiration (the PCC removes old notifications).
// FIXME use the WS and the requestId
func checkGenericEvent(timeout time.Duration, id uint64, str2check string, from *time.Time) (found bool, err error) {
	fmt.Printf("Looking for event [%s] %d from: %+v\n", str2check, id, from)
	start := time.Now()
	for time.Since(start) < timeout {
		if events, err := getEvents(); err == nil {
			for i := 0; i < len(events); i++ {
				event := events[i]
				//if from != nil && event.CreatedAt < ConvertToMillis(*from) {
				//	continue
				//}
				if event.TargetId == id && strings.Contains(strings.ToLower(events[i].Message), strings.ToLower(str2check)) {
					return true, nil
				}
			}
			time.Sleep(FREQUENCY * time.Second)
		} else {
			return false, err
		}
	}
	return false, fmt.Errorf("timeout error [%s] for id %d and contnet %s", timeout.String(), id, str2check)
}

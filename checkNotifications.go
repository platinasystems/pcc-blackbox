package main

import (
	"fmt"
	"github.com/platinasystems/tiles/pccserver/models"
	"strings"
	"time"
)

const (
	FREQUENCY                            = 10   // FIXME move to props
	AGENT_TIMEOUT                        = 300  // FIXME move to props
	COLLECTOR_TIMEOUT                    = 300  // FIXME move to props
	LLDP_TIMEOUT                         = 300  // FIXME move to props
	MAAS_INSTALL_TIMEOUT                 = 900  // FIXME move to props
	K8S_INSTALL_TIMEOUT                  = 1800 // FIXME move to props
	PORTUS_TIMEOUT                       = 600  // FIXME move to props
	PXEBOOT_TIMEOUT                      = 400  // FIXME move to props
	PORTUS_NOTIFICATION                  = "PORTUS"
	COLLECTOR_NOTIFICATION               = "COLLECTOR"
	AGENT_NOTIFICATION                   = "AGENT"
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
			events, err = Pcc.GetEvents()
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

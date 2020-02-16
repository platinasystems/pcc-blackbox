package main

import (
	"fmt"
	"strings"
	"time"

	pcc "github.com/platinasystems/pcc-blackbox/lib"
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
	PORTUS_NOTIFICATION                  = "[Portus] has been installed correctly"
	COLLECTOR_NOTIFICATION               = "The collector has been installed"
	AGENT_NOTIFICATION                   = "The agent has been installed"
	LLDP_NOTIFICATION                    = "[LLDPD] Installed version"
	PXEBOOT_NODE_ADD_NOTIFICATION        = "new node added successfully"
	PXEBOOT_NODE_ADD_FAILED_NOTIFICATION = "add node at  failed"
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
				events []pcc.Notification
				err    error
			)
			events, err = Pcc.GetNotifications()
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

	s.msg = fmt.Sprintf("\"%v\" notification not found in events..timeout exceeded", str2Check)
	s.isError = true
	found <- s
}

func checkGenericInstallation(id uint64, appTimeout time.Duration, str2Check string, from time.Time) (found bool, err error) {
	var (
		start   = time.Now()
		timeout = appTimeout * time.Second
		events  []pcc.Notification
	)
	found = false
	events, err = Pcc.GetNotifications()
	if err != nil {
		return
	}

	if err != nil {
		return
	} else if found {
		return
	}

	return checkingLoop(start, timeout, id, str2Check, events, from)
}

func checkMAASInstallation(id uint64, from time.Time) (found bool, err error) {
	var (
		start   = time.Now()
		timeout = MAAS_INSTALL_TIMEOUT * time.Second
		events  []pcc.Notification
	)
	found, err = checkingLoop(start, timeout, id, "[MAAS] Starting Bare-metal Role ", events, from)
	if err != nil {
		return false, fmt.Errorf("MAAS Step#1 - Timeout error [%d]", timeout)
	}
	found, err = checkingLoop(start, timeout, id, "Bare Metal Dependencies in progress", events, from)
	if err != nil {
		return false, fmt.Errorf("MAAS Step#2 - Timeout error [%d]", timeout)
	}
	found, err = checkingLoop(start, timeout, id, "Bare Metal Dependencies playbook completed", events, from)
	if err != nil {
		return false, fmt.Errorf("MAAS Step#3 - Timeout error [%d]", timeout)
	}
	found, err = checkingLoop(start, timeout, id, "Bare Metal Image Repository in progress", events, from)
	if err != nil {
		return false, fmt.Errorf("MAAS Step#4 - Timeout error [%d]", timeout)
	}
	found, err = checkingLoop(start, timeout, id, "Updated Platina Utility Linux source media", events, from)
	if err != nil {
		return false, fmt.Errorf("MAAS Step#5 - Timeout error [%d]", timeout)
	}
	found, err = checkingLoop(start, timeout, id, "Bare Metal Image Repository playbook completed", events, from)
	if err != nil {
		return false, fmt.Errorf("MAAS Step#6 - Timeout error [%d]", timeout)
	}
	found, err = checkingLoop(start, timeout, id, "[MAAS] Bare-metal deployment Role has been installed", events, from)
	if err != nil {
		return false, fmt.Errorf("MAAS Step#7 - Timeout error [%d]", timeout)
	}
	found, err = checkingLoop(start, timeout, id, "Bare Metal Multitenancy in progress", events, from)
	if err != nil {
		return false, fmt.Errorf("MAAS Step#8 - Timeout error [%d]", timeout)
	}
	found, err = checkingLoop(start, timeout, id, "Updating private deployment repository for tenant 'ROOT'", events, from)
	if err != nil {
		return false, fmt.Errorf("MAAS Step#9 - Timeout error [%d]", timeout)
	}
	found, err = checkingLoop(start, timeout, id, "Bare Metal Multitenancy playbook completed", events, from)
	if err != nil {
		return false, fmt.Errorf("MAAS Step#10 - Timeout error [%d]", timeout)
	}
	return found, nil
}

func checkGenericEvent(nodeId uint64, from time.Time, str2Check string, events []pcc.Notification, checkFrom bool) (found bool, err error) {
	found = false
	for i := 0; i < len(events); i++ {
		if checkFrom {
			if events[i].CreatedAt < ConvertToMillis(from) {
				continue
			}
		}
		if events[i].TargetId == nodeId {
			if events[i].Level == "error" {
				err = fmt.Errorf("%v", events[i].Message)
				fmt.Printf("Error event: [%v]\n", events[i])
				return
			}
			if strings.Contains(events[i].Message, str2Check) {
				found = true
				return
			}
		}
	}
	return
}

func checkingLoop(start time.Time, timeout time.Duration, id uint64, str2check string, events []pcc.Notification, from time.Time) (found bool, err error) {
	found = false
	for time.Since(start) < timeout {
		time.Sleep(FREQUENCY * time.Second)
		events, err = Pcc.GetNotifications()
		if err != nil {
			err = fmt.Errorf("getNofications error: %v", err)
			return
		}
		found, err = checkGenericEvent(id, from, str2check, events, true)
		if err != nil {
			return
		} else if found {
			return
		}
	}
	fmt.Printf("timeout [%v] [%v]\n", start, timeout)
	err = fmt.Errorf("Timeout error [%d]", timeout)
	return
}

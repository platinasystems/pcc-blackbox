package main

import (
	"encoding/json"
	"fmt"
	"github.com/platinasystems/tiles/pccserver/models"
	"strings"
	"time"
)

const (
	FREQUENCY            = 10
	AGENT_TIMEOUT        = 150
	COLLECTOR_TIMEOUT    = 150
	LLDP_TIMEOUT         = 300
	MAAS_INSTALL_TIMEOUT = 300
	K8S_INSTALL_TIMEOUT  = 1800
	PORTUS_TIMEOUT      = 400
	PXEBOOT_TIMEOUT = 400
	PORTUS_NOTIFICATION = "[Portus] has been installed correctly"
	COLLECTOR_NOTIFICATION = "The collector has been installed"
	AGENT_NOTIFICATION = "The agent has been installed"
	LLDP_NOTIFICATION = "[LLDPD] Installed version"
	PXEBOOT_NODE_ADD_NOTIFICATION = "new node added successfully"
	PXEBOOT_NODE_ADD_FAILED_NOTIFICATION = "add node at  failed"
)

type status struct {
	msg string
	isError bool
}
// Synchronize checking for installation
func syncCheckGenericInstallation(id uint64, appTimeout time.Duration, str2Check string, from time.Time, found chan status, breakLoop chan bool) {
	s := status{}
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
				s.msg = fmt.Sprintf("failed to getEvents ERROR: %v", err)
				s.isError = true
				found <- s
			}else {
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

func getEvents() (events []models.Notification, err error) {
	var (
		resp HttpResp
	)
	resp, _, err = pccGateway("GET", "pccserver/notifications/history?page=0&limit=50", nil)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(resp.Data, &events)
	if err != nil {
		return nil, err
	}
	return events, nil
}

//func checkAgentInstallation(id uint64) (found bool, err error) {
//	var (
//		start     = time.Now()
//		timeout   = AGENT_TIMEOUT * time.Second
//		events    []models.Notification
//		str2Check = "The agent has been installed"
//	)
//	events, err = getEvents()
//	if err != nil {
//		return false, err
//	}
//	if checkGenericEvent(id, start, str2Check, events, false) {
//		return true, nil
//	}
//	return checkingLoop(start, timeout, id, str2Check, events)
//}

//func checkCollectorInstallation(id uint64) (found bool, err error) {
//	var (
//		start     = time.Now()
//		timeout   = COLLECTOR_TIMEOUT * time.Second
//		events    []models.Notification
//		str2Check = "The collector has been installed"
//	)
//
//	events, err = getEvents()
//	if err != nil {
//		return false, err
//	}
//	if checkGenericEvent(id, start, str2Check, events, false) {
//		return true, nil
//	}
//	return checkingLoop(start, timeout, id, str2Check, events)
//}

//func checkLLDPInstallation(id uint64) (found bool, err error) {
//	var (
//		start     = time.Now()
//		timeout   = LLDP_TIMEOUT * time.Second
//		events    []models.Notification
//		str2Check = "[LLDPD] Installed version"
//	)
//	events, err = getEvents()
//	if err != nil {
//		return false, err
//	}
//	if checkGenericEvent(id, start, str2Check, events, false) {
//		return true, nil
//	}
//	return checkingLoop(start, timeout, id, str2Check, events)
//}

func checkGenericInstallation(id uint64, appTimeout time.Duration, str2Check string, from time.Time) (found bool, err error) {
	var (
		start     = time.Now()
		timeout   = appTimeout * time.Second
		events    []models.Notification
	)
	events, err = getEvents()
	if err != nil {
		return false, err
	}

	if checkGenericEvent(id, from, str2Check, events, true) {
				return true, nil
	}

	return checkingLoop(start, timeout, id, str2Check, events, from)
}

func checkMAASInstallation(id uint64, from time.Time) (found bool, err error) {
	var (
		start   = time.Now()
		timeout = MAAS_INSTALL_TIMEOUT * time.Second
		events  []models.Notification
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

func lightCheckMAASInstallation(id uint64, from time.Time) (found bool, err error) {
	var (
		start   = time.Now()
		timeout = MAAS_INSTALL_TIMEOUT * time.Second
		events  []models.Notification
	)
	found, err = checkingLoop(start, timeout, id, "[MAAS] Starting Bare-metal Role ", events, from)
	if err != nil {
		return false, fmt.Errorf("MAAS Step#1 - Timeout error [%d]", timeout)
	}
	found, err = checkingLoop(start, timeout, id, "[MAAS] Bare-metal deployment Role has been installed", events, from)
	if err != nil {
		return false, fmt.Errorf("MAAS Step#2 - Timeout error [%d]", timeout)
	}

	return found, nil
}

func checkGenericEvent(nodeId uint64, from time.Time, str2Check string, events []models.Notification, checkFrom bool) (found bool) {
	for i := 0; i < len(events); i++ {
		if checkFrom {
			if events[i].CreatedAt < ConvertToMillis(from) {
				continue
			}
		}
		if events[i].TargetId == nodeId {
			if strings.Contains(events[i].Message, str2Check) {
				return true
			}
		}
	}
	return false
}

func checkingLoop(start time.Time, timeout time.Duration, id uint64, str2check string, events []models.Notification, from time.Time) (found bool, err error) {
	for time.Since(start) < timeout {
		time.Sleep(FREQUENCY * time.Second)
		events, err = getEvents()
		if err != nil {
			return false, err
		}
		if checkGenericEvent(id, from, str2check, events, true) {
			return true, nil
		}
	}
	return false, fmt.Errorf("Timeout error [%d]", timeout)
}
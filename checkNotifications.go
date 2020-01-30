package main

import (
	"fmt"
	"strings"
	"time"

	pcc "github.com/platinasystems/pcc-blackbox/lib"
)

const (
	FREQUENCY              = 10
	AGENT_TIMEOUT          = 150
	COLLECTOR_TIMEOUT      = 150
	LLDP_TIMEOUT           = 300
	MAAS_INSTALL_TIMEOUT   = 300
	K8S_INSTALL_TIMEOUT    = 1800
	PORTUS_TIMEOUT         = 400
	PORTUS_NOTIFICATION    = "[Portus] has been installed correctly"
	COLLECTOR_NOTIFICATION = "The collector has been installed"
	AGENT_NOTIFICATION     = "The agent has been installed"
	LLDP_NOTIFICATION      = "[LLDPD] Installed version"
)

func checkGenericInstallation(id uint64, appTimeout time.Duration, str2Check string, from time.Time) (found bool, err error) {
	var (
		start   = time.Now()
		timeout = appTimeout * time.Second
		events  []pcc.Notification
	)
	events, err = Pcc.GetNotifications()
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

func checkGenericEvent(nodeId uint64, from time.Time, str2Check string, events []pcc.Notification, checkFrom bool) (found bool) {
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

func checkingLoop(start time.Time, timeout time.Duration, id uint64, str2check string, events []pcc.Notification, from time.Time) (found bool, err error) {
	for time.Since(start) < timeout {
		time.Sleep(FREQUENCY * time.Second)
		events, err = Pcc.GetNotifications()
		if err != nil {
			return false, err
		}
		if checkGenericEvent(id, from, str2check, events, true) {
			return true, nil
		}
	}
	return false, fmt.Errorf("Timeout error [%d]", timeout)
}

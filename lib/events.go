package pcc

import (
	"fmt"
	"github.com/platinasystems/tiles/pccserver/models"
	"strings"
	"time"
)

const (
	FREQUENCY = 10
)

type Status struct {
	Msg     string
	IsError bool
}

// For logging the notifications to stdout
// map[<notification_substring>]<do or don't terminate after notification is found>
type EventsToCheck map[string]bool

func (client *PccClient) GetEvents() (events []models.Notification, err error) {
	err = client.Get("pccserver/notifications/history", &events)
	return
}

func (client *PccClient) WaitForEvent(timeout time.Duration, targetId uint64, str2check string, requestId string) (found bool, err error) {
	fmt.Printf("Looking for event [%s] %d\n", str2check, targetId)
	start := time.Now()
	for time.Since(start) < timeout {
		if events, err := client.GetEvents(); err == nil {
			for i := 0; i < len(events); i++ {
				event := events[i]
				//if from != nil && event.CreatedAt < ConvertToMillis(*from) {
				//	continue
				//}
				if requestId != "" { // check by requestId
					if event.RequestId == requestId {
						found = true
						if event.Level == "error" {
							err = fmt.Errorf("error in event. %v", event)
						}
						return found, err
					}
				} else if event.CreatedAt < ConvertToMillis(start) {
					continue
				} else if event.TargetId == targetId && strings.Contains(strings.ToLower(events[i].Message), strings.ToLower(str2check)) {
					return true, nil
				}
			}
			time.Sleep(FREQUENCY * time.Second)
		} else {
			return false, err
		}
	}
	return false, fmt.Errorf("timeout error [%s] for id %d and contnet %s", timeout.String(), targetId, str2check)
}

func (client *PccClient) WaitForInstallation(id uint64, appTimeout time.Duration, app string, requestId string) (found bool, err error) {
	var (
		timeout = appTimeout * time.Second
	)
	str2Check := fmt.Sprintf("[%s] has been installed", app)
	return client.WaitForEvent(timeout, id, str2Check, requestId)
}

type Verifier interface {
	GetTimeout() time.Duration
	GetEventsToCheck() EventsToCheck
	GetBreakLoopChan() chan bool
}

//FIXME Verify is similar with checkNotifications.go
func (p *PccClient) Verify(startTime time.Time, v Verifier) (s Status) {
	done := make(chan Status)
	go p.SyncCheckGenericInstallation(0, v, startTime, done)
	s = <-done
	go func() {
		if v.GetBreakLoopChan() != nil {
			v.GetBreakLoopChan() <- true
		}
	}()

	return
}

// Synchronize checking for installation
//FIXME SyncCheckGenericInstallation is similar with checkNotifications.go
func (p *PccClient) SyncCheckGenericInstallation(id uint64, v Verifier, from time.Time, found chan Status) {
	s := Status{}
	timeout := v.GetTimeout() * time.Second
	for time.Since(from) < timeout {
		select {
		case <-v.GetBreakLoopChan():
			return
		default:
			var (
				events []Notification
				err    error
			)
			events, err = p.GetNotifications()
			if err != nil {
				s.Msg = fmt.Sprintf("failed to get notifications ERROR: %v", err)
				s.IsError = true
				found <- s
			} else {
			eventsLoop:
				for i := 0; i < len(events); i++ {
					if events[i].CreatedAt < ConvertToMillis(from) {
						continue
					}
					eventsToCheck := v.GetEventsToCheck()
					for msg, terminate := range eventsToCheck {
						eventMsg := events[i].Message
						if strings.Contains(eventMsg, msg) {
							if terminate {
								s.Msg = fmt.Sprintf("\"%v\" notification found in events", eventMsg)
								s.IsError = false
								found <- s
								break eventsLoop
							} else {
								fmt.Println("notification: ", eventMsg)
								delete(eventsToCheck, msg)
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

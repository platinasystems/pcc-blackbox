package pcc

import (
	"fmt"
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

type Verifier interface {
	GetTimeout() time.Duration
	GetEventsToCheck() EventsToCheck
	GetBreakLoopChan() chan bool
}

func (p *PccClient) Verify(startTime time.Time, v Verifier) (s Status) {
	done := make(chan Status)
	go p.syncCheckGenericInstallation(0, v, startTime, done)
	s = <-done
	go func() {
		if v.GetBreakLoopChan() != nil {
			v.GetBreakLoopChan() <- true
		}
	}()

	return
}

// Synchronize checking for installation
func (p *PccClient) syncCheckGenericInstallation(id uint64, v Verifier, from time.Time, found chan Status) {
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

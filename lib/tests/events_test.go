package mailer

import (
	"fmt"
	"github.com/platinasystems/go-common/http"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/tiles/pccserver/models"
	"testing"
	"time"
)

const EVENTOK = "event message"
const OLD = 1585299686333

var n = 1

type FakeRestClient struct {
}

// GET
func (frc *FakeRestClient) Get(endPoint string, out interface{}) (err error) {
	// Will return the correct event at the 3rd cycle
	var e []models.Notification
	fmt.Println("Fake call to", endPoint)
	var event models.Notification
	switch n {
	case 1:
		event = models.Notification{CreatedAt: OLD, TargetId: 1, Message: "aaaaaaa"}
		e = append(e, event)
		e = append(e, event)
		event = models.Notification{CreatedAt: OLD, TargetId: 1, Message: "bbbbbbb"}
	case 2:
		event = models.Notification{CreatedAt: OLD, TargetId: 1, Message: "aaaaaaa"}
		e = append(e, event)
		event = models.Notification{CreatedAt: OLD, TargetId: 1, Message: EVENTOK}
		e = append(e, event)
		e = append(e, event)
	case 3:
		event = models.Notification{CreatedAt: OLD, TargetId: 1, Message: "aaaaaaa"}
		e = append(e, event)
		event = models.Notification{CreatedAt: OLD, TargetId: 1, Message: "ccccccc"}
		e = append(e, event)
		event = models.Notification{CreatedAt: pcc.ConvertToMillis(time.Now()), TargetId: 1, Message: EVENTOK}
		e = append(e, event)
	}
	events := out.(*[]models.Notification)
	*events = e
	n = n + 1
	return
}

// DELETE
func (pcc *FakeRestClient) Delete(endPoint string, out interface{}) (err error) {
	return // Nothing to do
}

// POST
func (pcc *FakeRestClient) Post(endPoint string, data interface{}, out interface{}) (err error) {
	return // Nothing to do
}

// PUT
func (pcc *FakeRestClient) Put(endPoint string, data interface{}, out interface{}) (err error) {
	return // Nothing to do
}

// PUT
func (pcc *FakeRestClient) PutFile(endPoint string, filePath string, fields map[string]string, out interface{}) (err error) { // FIXME all services should share the same structure
	return // Nothing to do
}

func (pcc *FakeRestClient) GetFile(endPoint string) (content string, err error) {
	return // Nothing to do
}

func TestWaitForEvent(t *testing.T) {
	var (
		rc  http.PlatinaRestService
		krc FakeRestClient
		pcc pcc.PccClient
	)
	rc = &krc
	pcc.RestClient = &rc

	if _, err := pcc.WaitForEvent(time.Duration(180)*time.Second, 1, EVENTOK, "", nil); err != nil {
		t.Fatal(err)
	}
}

package pcc

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Rest

type Monitor struct {
	Token    string
	Url      string
	Insecure bool

	Ws Ws
}

func (m *Monitor) GetHistorical(topic string, from int64, to int64, nodeIds []uint64, fields []string) (string, error) {

	client := m.getTLSClient()
	endpoint := fmt.Sprintf("monitor/topic/%s/historical", topic)

	// Prepare body content
	content := getHistorical(from, to, nodeIds, fields)

	// Prepare post
	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s/%s", m.Url, endpoint), strings.NewReader(content))
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", m.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return "", err
	}

	body, _ := ioutil.ReadAll(resp.Body)
	return string(body), nil
}

func (m *Monitor) getTLSClient() *http.Client {

	// Get the SystemCertPool, continue with an empty pool on error
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: m.Insecure, RootCAs: rootCAs},
	}

	return &http.Client{Transport: tr}
}

func getHistorical(from int64, to int64, nodeIds []uint64, fields []string) string {

	type TimeRange struct {
		From int64 `json:"from"`
		To   int64 `json:"to"`
	}

	type Historical struct {
		TimeRange TimeRange `json:"timeRange"`
		NodeIds   []uint64  `json:"nodeIds"`
		Fields    []string  `json:"fields"`
	}
	h := Historical{TimeRange: TimeRange{From: from, To: to}, NodeIds: nodeIds, Fields: fields}
	b, err := json.Marshal(h)

	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

// Websocket

type Ws struct {
	Connection *websocket.Conn
	Error      error
	Done       chan struct{}
	Statistics *WsStatistics
}

type WsData struct {
	Message string        `json:"message"`
	Content WsDataContent `json:"content"`
}

type WsDataContent struct {
	Resource  string        `json:"resource"`
	Timestamp uint64        `json:"timestamp"`
	NodeId    uint64        `json:"nodeId"`
	NodeUuid  string        `json:"nodeUuid"`
	Node      string        `json:"node"`
	Metrics   []interface{} `json:"metrics"`
}

type WsStatistics struct {
	ActiveNodeMap map[uint64]uint64
}

func NewMonitor(token string, url string, insecure bool) *Monitor {
	m := &Monitor{Token: token, Url: url, Insecure: insecure, Ws: Ws{Statistics: &WsStatistics{ActiveNodeMap: make(map[uint64]uint64)}}}
	return m
}

func (m *Monitor) WSConnect() bool {
	u := url.URL{Scheme: "wss", Host: m.Url, Path: "/monitor/data/live", RawQuery: fmt.Sprintf("Authorization=%s", strings.Replace(m.Token, "Bearer ", "", 1))}

	// Get the SystemCertPool, continue with an empty pool on error
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{RootCAs: rootCAs, InsecureSkipVerify: m.Insecure}

	m.Ws.Connection, _, m.Ws.Error = dialer.Dial(u.String(), nil)
	if m.Ws.Error != nil {
		return false
	}

	m.Ws.Done = make(chan struct{})
	return true
}

func (m *Monitor) WsSendingMetrics(metrics []string) {
	if m.Ws.Connection != nil {
		x := NewMetrics()
		for _, metric := range metrics {
			x.AddMetrics(metric)
		}
		m.Ws.Connection.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s", x.metricsToJson())))
	}
}

func (m *Monitor) WsListeningLoop() {
	if m.Ws.Connection != nil {
		go func() {
			defer close(m.Ws.Done)
			for {
				_, message, err := m.Ws.Connection.ReadMessage()
				if err != nil {
					return
				}
				m.wsParseMessage(message)
			}
		}()
	}
}

func (m *Monitor) WsReadTimeout(timeout time.Duration) {
	select {
	case <-m.Ws.Done:
		return
	case <-time.After(timeout):
		log.Println("interrupt")
		return
	}
}

func (m *Monitor) wsParseMessage(jsonMessage []byte) bool {
	d := &WsData{}
	err := json.Unmarshal(jsonMessage, d)
	if err != nil {
		log.Println("Data not recognized! ", err)
		return false
	}

	// Get topic
	s := strings.Index(d.Message, "[")
	if s == -1 {
		return false
	}
	s += len("[")
	e := strings.Index(d.Message, "]")
	if e == -1 {
		return false
	}
	//topic := d.Message[s:e]

	m.Ws.Statistics.ActiveNodeMap[d.Content.NodeId] += 1

	return true
}

// Metrics

type Resource struct {
	Resource string `json:"resource"`
}

type Metrics struct {
	Metrics []Resource `json:"metrics"`
}

func NewMetrics() *Metrics {
	return &Metrics{Metrics: []Resource{}}
}

func (m *Metrics) AddMetrics(r string) *Metrics {
	m.Metrics = append(m.Metrics, Resource{Resource: r})
	return m
}

func (m *Metrics) metricsToJson() string {
	b, err := json.Marshal(&m)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(b)
}

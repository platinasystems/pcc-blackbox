package pcc

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"strings"
	"time"
)

// Rest

type HistoricalData struct {
	Resource                 string               `json:"resource"`
	Timestamp                string               `json:"timestamp"`
	NodeId                   uint64               `json:"nodeId"`
	NodeUuid                 string               `json:"nodeUuid"`
	Node                     string               `json:"node"`
	CompensatedFromTimestamp interface{}          `json:"compensatedFromTimestamp"`
	Metrics                  []map[string]float32 `json:"metrics"`
}

func (p *PccClient) GetHistoricalData(topic string, from int64, to int64, nodeIDs []uint64, fields []string) (result map[string][]HistoricalData, err error) {

	var (
		body []byte
		data []byte
		resp HttpResp
	)

	endpoint := fmt.Sprintf("monitor/topic/%s/historical", topic)

	data = toHistoricalDataRequest(from, to, nodeIDs, fields)
	if err != nil {
		return
	}

	resp, body, err = p.pccGateway("POST", endpoint, data)
	if err != nil {
		return
	}

	if resp.Status != 200 && resp.Status != 0 {
		err = fmt.Errorf("%v: %v", resp.Error, resp.Message)
		return
	}

	result = toHistoricalDataResponse(body)
	return
}

func toHistoricalDataRequest(from int64, to int64, nodeIDs []uint64, fields []string) []byte {

	type TimeRange struct {
		From int64 `json:"from"`
		To   int64 `json:"to"`
	}

	type Historical struct {
		TimeRange TimeRange `json:"timeRange"`
		NodeIds   []uint64  `json:"nodeIDs"`
		Fields    []string  `json:"fields"`
	}
	h := Historical{TimeRange: TimeRange{From: from, To: to}, NodeIds: nodeIDs, Fields: fields}
	b, err := json.Marshal(h)
	if err != nil {
		return nil
	}
	return b
}

func toHistoricalDataResponse(response []byte) map[string][]HistoricalData {
	data := make(map[string][]HistoricalData)
	json.Unmarshal(response, &data)
	return data
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

func (p *PccClient) WsLiveDataConnect() (ws *Ws) {

	ws = &Ws{Statistics: &WsStatistics{ActiveNodeMap: make(map[uint64]uint64)}}
	u := url.URL{Scheme: "wss", Host: p.pccIp + ":9999", Path: "/monitor/data/live", RawQuery: fmt.Sprintf("Authorization=%s", strings.Replace(p.bearer, "Bearer ", "", 1))}

	// Get the SystemCertPool, continue with an empty pool on error
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{RootCAs: rootCAs, InsecureSkipVerify: true}

	ws.Connection, _, ws.Error = dialer.Dial(u.String(), nil)
	if ws.Error != nil {
		return
	}

	ws.Done = make(chan struct{})
	return
}

func (ws *Ws) WsSendingMetrics(metrics []string) {
	if ws.Connection != nil {
		x := newMetrics()
		for _, metric := range metrics {
			x.addMetrics(metric)
		}
		ws.Connection.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s", x.metricsToJson())))
	}
}

func (ws *Ws) WsListen() {
	if ws.Connection != nil {
		go func() {
			defer close(ws.Done)
			for {
				_, message, err := ws.Connection.ReadMessage()
				if err != nil {
					return
				}
				ws.wsParseMessage(message)
			}
		}()
	}
}

func (ws *Ws) WsListenTimeout(timeout time.Duration) {
	select {
	case <-ws.Done:
		return
	case <-time.After(timeout):
		log.Println("interrupt")
		return
	}
}

func (ws *Ws) wsParseMessage(jsonMessage []byte) bool {
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

	ws.Statistics.ActiveNodeMap[d.Content.NodeId] += 1

	return true
}

// Metrics

type Resource struct {
	Resource string `json:"resource"`
}

type Metrics struct {
	Metrics []Resource `json:"metrics"`
}

func newMetrics() *Metrics {
	return &Metrics{Metrics: []Resource{}}
}

func (m *Metrics) addMetrics(r string) *Metrics {
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

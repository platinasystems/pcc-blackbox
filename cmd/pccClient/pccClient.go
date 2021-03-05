// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"github.com/platinasystems/go-common/http"
	"github.com/platinasystems/go-common/logs"

	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type Credential struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

type PccClient struct {
	pccIp      string
	bearer     string
	RestClient *http.PlatinaRestService
}

type TimeRange struct {
	From uint64 `json:"from"`
	To   uint64 `json:"to"`
}

type Historical struct {
	TimeRange TimeRange `json:"timeRange"`
	Nodes     []string  `json:"nodes"`
	Fields    []string  `json:"fields,omitempty"`
}

// key is nodeId but as string
type HistoricalOut map[string][]TimeSeriesElem

type TimeSeriesElem struct {
	Timestamp int64                    `json:"timestamp"`
	Metrics   []map[string]interface{} `json:"metrics,omitempty"`
}

func (pcc *PccClient) getClient() (rc *http.PlatinaRestService) {
	if pcc.RestClient == nil {
		var prs http.PlatinaRestService
		prc := http.PlatinaRestClient{Address: pcc.pccIp, Token: pcc.bearer, Port: 9999}
		prs = &prc
		rc = &prs
	} else {
		rc = pcc.RestClient
	}

	return
}

// GET
func (pcc *PccClient) Get(endPoint string, out interface{}, options ...interface{}) (err error) {
	err = (*pcc.getClient()).Get(endPoint, out, options...)
	return
}

// DELETE
func (pcc *PccClient) Delete(endPoint string, data interface{}, out interface{}, options ...interface{}) (err error) {
	err = (*pcc.getClient()).Delete(endPoint, data, out, options...)
	return
}

// POST
func (pcc *PccClient) Post(endPoint string, data interface{}, out interface{}, options ...interface{}) (err error) {
	err = (*pcc.getClient()).Post(endPoint, data, out, options...)
	return
}

// PUT
func (pcc *PccClient) Put(endPoint string, data interface{}, out interface{}, options ...interface{}) (err error) {
	err = (*pcc.getClient()).Put(endPoint, data, out, options...)
	return
}

// PUT one file (multipart/form-data)
func (pcc *PccClient) PutFile(endPoint string, filePath string, fields map[string]string, out interface{}) (err error) { // FIXME all services should share the same structure
	err = (*pcc.getClient()).PutFile(endPoint, filePath, fields, out)
	return
}

// PUT one or more files (multipart/form-data)
func (pcc *PccClient) PutFiles(method string, endPoint string, files map[string]string, fields map[string]string, out interface{}) (err error) { // FIXME all services should share the same structure
	err = (*pcc.getClient()).PutFiles(method, endPoint, files, fields, out)
	return
}

func (pcc *PccClient) GetFile(endPoint string) (content string, err error) {
	content, err = (*pcc.getClient()).GetFile(endPoint)
	return
}

func Authenticate(PccIp string, cred Credential) (pcc *PccClient, err error) {
	var out struct{ Token string }
	rc := http.PlatinaRestClient{Address: PccIp, Port: 9999}
	client := PccClient{pccIp: PccIp}
	if err = rc.Post("security/auth", cred, &out); err == nil {
		client.bearer = fmt.Sprintf("Bearer %s", out.Token)
	}
	pcc = &client
	return
}

func pPrint(out interface{}) {
	switch t := out.(type) {
	case []interface{}:
		j, _ := json.MarshalIndent(t, "", " ")
		fmt.Println(string(j))
	case map[string]interface{}:
		j, _ := json.MarshalIndent(t, "", " ")
		fmt.Println(string(j))
	default:
		fmt.Println(out)
	}
}

func pPrintJ(out interface{}) {
	s := fmt.Sprintf("%v", out)
	switch {
	case strings.HasPrefix(s, "{"):
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(s), &m); err == nil {
			j, _ := json.MarshalIndent(m, "", " ")
			fmt.Println(string(j))
			return
		}
	case strings.HasPrefix(s, "["):
		var a []interface{}
		if err := json.Unmarshal([]byte(s), &a); err == nil {
			j, _ := json.MarshalIndent(a, "", " ")
			fmt.Println(string(j))
			return
		} else {
			fmt.Println(err)
		}
	}
	fmt.Println(out)
}

func (d HistoricalOut) Print(nodes, fields []string) {
	var all bool
	fields = append([]string{"hostname"}, fields...)
	if len(nodes) == 0 {
		all = true
	}
	nodeMap := map[string]bool{}
	for _, node := range nodes {
		nodeMap[node] = true
	}
	for nodeId, ts := range d {
		for i, t := range ts {
			if len(t.Metrics) == 0 {
				continue
			}
			header := fmt.Sprintf("%-10v%30v", "nodeId", "timestamp")
			if i == 0 {
				for _, field := range fields {
					if _, ok := t.Metrics[0][field]; ok {
						header += fmt.Sprintf("%20v", field)
					}
				}
			}

			zone, _ := time.LoadLocation("America/Los_Angeles")
			timestamp := time.Unix(0, t.Timestamp*1000000).In(zone).Format("MST 2006-01-02 15:04:05.000")
			line := fmt.Sprintf("%-10v%30v", nodeId, timestamp)
			include := nodeMap[fmt.Sprintf("%v", nodeId)] || all
			for _, field := range fields {
				if value, ok := t.Metrics[0][field]; ok {
					line += fmt.Sprintf("%20v", value)
					if field == "hostname" && nodeMap[fmt.Sprintf("%v", value)] {
						include = true
					}
				}
			}
			if include {
				if i == 0 {
					fmt.Println(header)
				}
				fmt.Println(line)
			}
		}
	}
}

func main() {
	var (
		nodes, fields string
		raw           bool
	)
	cred := Credential{
		UserName: "admin",
		Password: "admin",
	}
	usage := "[ip addr|domain name] [endpoint] [GET|POST|PUT] [data]"
	usage2 := "[ip addr|domain name] history [topic] [t1] [t2] [-n nodes] [-f fields]"
	example := "172.17.2.238 history summary 0 30m -n \"i60\" -f \"cpuLoad realUsedMem inodeUsage networkThrought\""
	if len(os.Args) < 4 {
		fmt.Println("usage1", os.Args[0], usage)
		fmt.Println("usage2", os.Args[0], usage2)
		fmt.Println("example", os.Args[0], example)
		return
	}
	if strings.EqualFold(os.Args[2], "history") && len(os.Args) < 6 {
		fmt.Println("usage2", os.Args[0], usage2)
		fmt.Println("example", os.Args[0], example)
		return
	}

	for i, arg := range os.Args {
		if arg == "--raw" {
			raw = true
		}
		if len(os.Args) <= i+1 {
			break
		}
		p := os.Args[i+1]
		switch arg {
		case "-n":
			nodes = p
		case "-f":
			fields = p
		case "-u":
			cred.UserName = p
		case "-p":
			cred.Password = p
		}

	}

	log.InitWithDefault(nil)
	addr := os.Args[1]
	endpoint := os.Args[2]
	cmd := os.Args[3]
	hasData := len(os.Args) >= 5
	data := "{}"
	if hasData {
		data = os.Args[4]
	}
	pcc, err := Authenticate(addr, cred)
	if err != nil {
		fmt.Println("Authentication error:", err)
		return
	}

	var out interface{}
	switch {
	case strings.EqualFold(endpoint, "history"):
		var out2 HistoricalOut
		topic := os.Args[3]
		now := time.Now()
		t1, errt1 := time.ParseDuration(os.Args[4])
		t2, errt2 := time.ParseDuration(os.Args[5])
		if errt1 != nil || errt2 != nil {
			fmt.Println(usage2)
			fmt.Println("t1 and t2 are time from now, e.g. 1h, 3m, etc.")
			return
		}
		endpoint = fmt.Sprintf("monitor/topic/%v/historical", topic)
		start := t1
		stop := t2
		if t1 < t2 {
			start = t2
			stop = t1
		}
		timeRange := TimeRange{
			From: uint64(now.Add(-start).Unix()) * 1000,
			To:   uint64(now.Add(-stop).Unix()) * 1000,
		}
		d := Historical{
			TimeRange: timeRange,
			Nodes:     strings.Fields(nodes), // FIXME seems to get all the nodes anyway
			//Fields:    strings.Fields(fields),
		}
		if raw {
			pcc.Post(endpoint, &d, &out)
		} else {
			pcc.Post(endpoint, &d, &out2)
			out2.Print(strings.Fields(nodes), strings.Fields(fields))
			return
		}
	case strings.EqualFold(cmd, "get"):
		pcc.Get(endpoint, &out, nil)
	case strings.EqualFold(cmd, "post"):
		pcc.Post(endpoint, &data, &out)
	case strings.EqualFold(cmd, "put"):
	}

	pPrint(out)
}

package pcc

import "github.com/prometheus/common/model"

func (pcc *PccClient) InstantQuery(query string) (data *model.Sample, err error) {
	endpoint := "platina-monitor/monitor/query"
	request := map[string]interface{}{"query": query}
	var m model.Sample
	err = pcc.Post(endpoint, &request, &m)
	data = &m
	return
}

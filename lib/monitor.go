// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"fmt"
)

////
// Monitor Rest Client
////

// Get history
func (pcc *PccClient) GetHistory(topic string, from int64, to int64) (data *map[string][]interface{}, err error) {
	request := make(map[string]map[string]int64)
	request["timeRange"] = map[string]int64{"from": from, "to": to}

	m := make(map[string][]interface{})
	err = pcc.Post(fmt.Sprintf("monitor/topic/%s/historical", topic), &request, &m)
	data = &m
	return
}

func (pcc *PccClient) GetNodesHistory(topic string, from uint64, to uint64, nodeIDs []uint64, fields []string) (data *map[string]interface{}, err error) {
	endpoint := fmt.Sprintf("monitor/topic/%s/historical", topic)
	request := map[string]interface{}{"fields": fields, "nodeIDs": nodeIDs, "timeRange": map[string]uint64{"from": from, "to": to}}
	m := make(map[string]interface{})
	err = pcc.Post(endpoint, &request, &m)
	data = &m
	return
}

// Return last sample
func (pcc *PccClient) GetLiveSample(topic string, nodeId uint64) (data *map[string]interface{}, err error) {
	m := make(map[string]interface{})
	err = pcc.Get(fmt.Sprintf("monitor/topic/%s/live/%d", topic, nodeId), &m)
	data = &m
	return
}

// List topic
func (pcc *PccClient) GetTopics() (data []string, err error) {
	err = pcc.Get("monitor/topic", &data)
	return
}

// List schema for topic
func (pcc *PccClient) GetSchema(topic string) (data map[int64]interface{}, err error) {
	err = pcc.Get(fmt.Sprintf("monitor/topic/%s", topic), &data)
	return
}

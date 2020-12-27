package main

import (
	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"
)

var (
	envFile string = "testEnv.json"
	Pcc     *pcc.PccClient
	Env     testEnv
)

type testEnv struct {
	Env                   string
	PccIp                 string
	DockerStats           pcc.DockerStatsConfig
	AuthenticationProfile pcc.AuthenticationProfile
	PortusConfiguration   pcc.PortusConfiguration
	DBConfiguration       *pcc.DBConfiguration
	SshConfiguration      *pcc.SshConfiguration
	CephConfiguration     pcc.CephConfiguration
	K8sAppConfiguration   pcc.K8sAppConfiguration
}

func main() {
	var (
		ssh              pcc.SSHHandler
		agentBuildId     string
		collectorBuildId string
		wg               sync.WaitGroup
		r                sync.Map
	)

	data, err := ioutil.ReadFile(envFile)
	if err != nil {
		panic(fmt.Errorf("Error opening %v: %v", envFile, err))
	}

	if err = json.Unmarshal(data, &Env); err != nil {
		if jsonErr, ok := err.(*json.SyntaxError); ok {
			// emacs users can use M-x goto-char <offset>
			part := data[jsonErr.Offset-10 : jsonErr.Offset+10]
			err = fmt.Errorf("%w ~ error near '%s' (offset %d)",
				err, part, jsonErr.Offset)
		}
		panic(fmt.Errorf("error unmarshalling %v: %v\n",
			envFile, err.Error()))
	}

	pcc.InitDB(Env.DBConfiguration)   // Init the DB handler
	pcc.InitSSH(Env.SshConfiguration) // Init the SSH handler

	log.InitWithDefault(nil)

	credential := pcc.Credential{ // FIXME move to json
		UserName: "admin",
		Password: "admin",
	}
	if Pcc, err = pcc.Authenticate(Env.PccIp, credential); err != nil {
		panic(fmt.Errorf("Authentication error: %v\n", err))
	}

	apps, err := Pcc.GetApps()

	for _, app := range apps {
		var m map[string]interface{}
		if app.Name == "agent" {
			json.Unmarshal([]byte(app.Metadata), &m)
			agentBuildId = strings.TrimSpace(fmt.Sprintf("%v", m["BuildId"]))
		}
		if app.Name == "collector" {
			json.Unmarshal([]byte(app.Metadata), &m)
			collectorBuildId = strings.TrimSpace(fmt.Sprintf("%v", m["BuildId"]))
		}
	}

	fmt.Println("expected agent buildid", agentBuildId)
	fmt.Println("expected collector buildid", collectorBuildId)

	timeOut := 10 * time.Second
	if nodes, err := Pcc.GetNodes(); err == nil {
		for _, node := range *nodes {
			wg.Add(1)
			hostIp := node.Host
			nodeId := node.Id
			nodeName := node.Name
			go func() {
				defer func() {
					wg.Done()
				}()
				var agentReady, collectorReady bool
				start := time.Now()
				for {
					if time.Since(start) > timeOut {
						break
					}
					if stdO, _, err := ssh.Run(hostIp,
						"/opt/platina/pcc/bin/pccagent daemon buildid", "-q"); err == nil {
						buildid := strings.TrimSpace(stdO)
						match := buildid == agentBuildId
						if match {
							stdO, _, _ := ssh.Run(hostIp,
								"sudo /opt/platina/pcc/bin/pccagent package agent", "-q")
							lines := strings.Split(stdO, "\n")
							OK := false
							hasDesired := false
							for _, line := range lines {
								if strings.Contains(line, "Desired Iface Config Received") &&
									strings.Contains(line, "true") {
									hasDesired = true
								}
								if strings.Contains(line, "\"OK\"") && strings.Contains(line, "true") {
									OK = true
								}
							}
							if OK && hasDesired {
								agentReady = true
								//fmt.Printf("%v agent ready\n", nodeName)
							}
						}
					}
					if stdO, _, err := ssh.Run(hostIp,
						"/opt/platina/pcc/bin/systemCollector daemon buildid", "-q"); err == nil {
						buildid := strings.TrimSpace(stdO)
						if buildid == collectorBuildId {
							collectorReady = true
							//fmt.Printf("%v collector ready\n", nodeName)
						}
					}
					key := fmt.Sprintf("%v(id:%v)", nodeName, nodeId)
					r.Store(key, agentReady && collectorReady)
					if agentReady && collectorReady {
						fmt.Println(key, "ready")
						break
					} else {
						time.Sleep(2 * time.Second)
					}
				}
			}()
		}
		wg.Wait()
		r.Range(func(id, v interface{}) bool {
			p := v.(bool)
			if !p {
				fmt.Println(id, "failed to come up in", timeOut)
			}
			return true
		})
	}
}

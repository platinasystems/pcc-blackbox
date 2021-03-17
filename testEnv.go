package main

import (
	"errors"
	"fmt"

	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
)

// Note: The lib directory is intended to be the start of an SDK, so
//       test configuration/infrastructure should NOT be in lib.
//       All the configuration below with "pcc." will eventually
//       removed from lib.  Please do not add more.

type testEnv struct {
	Env                   string
	PccIp                 string
	Invaders              []invader
	Servers               []server
	DockerStats           pcc.DockerStatsConfig
	AuthenticationProfile pcc.AuthenticationProfile
	PortusConfiguration   pcc.PortusConfiguration
	DBConfiguration       *pcc.DBConfiguration
	SshConfiguration      *pcc.SshConfiguration
	CephConfiguration     pcc.CephConfiguration
	K8sAppConfiguration   pcc.K8sAppConfiguration
	Availability          Availability
	NetIpam               []netIpam
	NetCluster            []netCluster
}

type node struct {
	Id            uint64
	HostIp        string
	BMCIp         string
	BMCUser       string
	BMCUsers      []string
	BMCPass       string
	KeyAlias      []string
	NetInterfaces []netInterface
}

type netInterface struct {
	Name         string
	Cidrs        []string
	Gateway      string
	MacAddr      string
	IsManagement bool
	ManagedByPcc bool
	AdminStatus  string
	Speed        string
	Autoneg      string
	Fec          string
	Media        string
	Mtu          string
}

type invader struct {
	node
}

type server struct {
	node
}

type Availability struct {
	FakeAddress  string
	Inaccessible string
}

type netCluster struct {
	Name        string
	ControlCIDR string
	DataCIDR    string
	IgwPolicy   string
	Nodes       []netNode
}

type netNode struct {
	IpAddr       string
	LocalAs      string
	BgpNeighbors []bgpPeer
}

type bgpPeer struct {
	NeighborIp string
	RemoteAs   string
}

type netIpam struct {
	Name      string
	Subnet    string
	PubAccess bool
	Routed    bool
}

var exampleEnv = testEnv{
	PccIp:        "172.17.2.238",
	Availability: Availability{"ImNotWorking", "172.17.0.0"},
	Invaders: []invader{
		invader{
			node{
				HostIp:   "172.17.2.60",
				BMCIp:    "172.17.3.60",
				BMCUser:  "ADMIN",
				BMCUsers: []string{"ADMIN"},
				BMCPass:  "ADMIN",
				NetInterfaces: []netInterface{
					netInterface{
						Name:  "eth0",
						Cidrs: []string{"172.17.2.60/23"},
					},
				},
			},
		},
	},
}

func (te *testEnv) setNodeId(host string, id uint64) (found bool) {
	found = false

	for i := range te.Invaders {
		if te.Invaders[i].HostIp == host {
			te.Invaders[i].Id = id
			found = true
			break
		}
	}

	if !found {
		for i := range te.Servers {
			if te.Servers[i].HostIp == host {
				te.Servers[i].Id = id
				break
			}
		}
	}
	return
}

func (te *testEnv) GetNodeByHost(host string) *node {
	for i := range te.Invaders {
		if te.Invaders[i].HostIp == host {
			return &te.Invaders[i].node
		}
	}

	for i := range te.Servers {
		if te.Servers[i].HostIp == host {
			return &te.Servers[i].node
		}
	}
	return nil
}

func (te *testEnv) IsNodeAlreadyAdded(host string) bool {
	if n := te.GetNodeByHost(host); n != nil {
		return n.Id > 0
	}

	return false
}

func (te *testEnv) CheckPccIp() (err error) {
	if te.PccIp == "" {
		err = errors.New("No PccIp configuration in Env file")
	}
	return
}
func (netInt *netInterface) CheckNetInt() (err error) {
	if netInt.AdminStatus != pcc.INTERFACE_STATUS_UP {
		log.AuctaLogger.Infof("CheckNetInt [%v] AdminStatus [%v]",
			netInt.Name, netInt.AdminStatus)
		return
	}
	switch netInt.Autoneg {
	case "true", "on":
		if netInt.Speed != "" {
			err = fmt.Errorf("Error in configurtion (speed must "+
				"not be set with autoneg off for [%v] speed "+
				"[%v] auto [%v]",
				netInt.Name, netInt.Speed, netInt.Autoneg)
		}
		return
	case "false", "off", "":
		if netInt.Speed == "" {
			err = fmt.Errorf("Error in configurtion (speed must "+
				"be set with autoneg on for [%v] speed [%v] auto [%v]",
				netInt.Name, netInt.Speed, netInt.Autoneg)
		}
		return
	}
	if netInt.MacAddr == "" || netInt.Mtu == "" {
		err = fmt.Errorf("Error in configuration parameters (MacAddr,"+
			" Mtu) for [%v]", netInt.Name)
		return
	}
	if !(len(netInt.Cidrs) > 0) {
		err = fmt.Errorf("There are no Cidrs in Env File for [%v]",
			netInt.Name)
		return
	}
	for _, cidr := range netInt.Cidrs {
		if cidr == "" {
			err = fmt.Errorf("Error in configuration parameters "+
				"(check: cidr) [%v]", netInt.Name)
			return
		}
	}
	return
}

func (node *node) CheckNode() (err error) {
	if node.HostIp == "" || node.BMCIp == "" || node.BMCUser == "" || node.BMCPass == "" {
		err = errors.New("Error in configuration parameters (check: HostIp, BMCIp, BMCUser, BMCPass)")
		return
	}
	if !(len(node.BMCUsers) > 0) {
		err = errors.New("There are no BMCUsers in Env File")
		return
	}
	for _, BMCUser := range node.BMCUsers {
		if BMCUser == "" {
			err = errors.New("Error in configuration parameters (check: BMCUser)")
			return
		}
	}
	if !(len(node.NetInterfaces) > 0) {
		err = errors.New("There are no NetInterfaces in Env File")
		return
	}
	for _, netInt := range node.NetInterfaces {
		err = netInt.CheckNetInt()
		if err != nil {
			return
		}
	}
	return
}

func (te *testEnv) CheckInvaders() (err error) {
	if !(len(te.Invaders) > 0) {
		err = errors.New("there are no invaders in Env file")
		return
	}
	for _, inv := range te.Invaders {
		err = inv.CheckNode()
		if err != nil {
			return
		}
	}
	return
}

func (te *testEnv) CheckServers() (err error) {
	if !(len(te.Servers) > 0) {
		err = errors.New("there are no servers in Env file")
		return
	}
	for _, serv := range te.Servers {
		err = serv.CheckNode()
		if err != nil {
			return
		}
	}
	return
}

func (ipam *netIpam) CheckNetIpam() (err error) {
	if ipam.Name == "" || ipam.Subnet == "" {
		err = errors.New("Error in configuration parameters (check: ipam Name, ipam Subnet)")
		return
	}
	return
}

func (te *testEnv) CheckNetIpams() (err error) {
	if !(len(te.NetIpam) > 0) {
		err = errors.New("there are no NetIpam in Env file")
		return
	}
	for _, ipam := range te.NetIpam {
		err = ipam.CheckNetIpam()
		if err != nil {
			return
		}
	}
	return
}

func (node *netNode) CheckNetNode() (err error) {
	if node.IpAddr == "" {
		err = errors.New("Error in configuration parameters (check: netNode IpAddr)")
		return
	}
	for _, bgpNeigh := range node.BgpNeighbors {
		if bgpNeigh.NeighborIp == "" || bgpNeigh.RemoteAs == "" {
			err = errors.New("Error in configuration parameters (check: NeighborIp, RemoteAs)")
			return
		}
	}
	return
}

func (cluster *netCluster) CheckNetCluster() (err error) {
	if cluster.Name == "" || cluster.ControlCIDR == "" || cluster.DataCIDR == "" || cluster.IgwPolicy == "" {
		err = errors.New("Error in configuration parameters (check: netCluster Name, netCluster ControlCIDR, netCluster DataCIDR, netCluster IgwPolicy)")
		return
	}
	if !(len(cluster.Nodes) > 0) {
		err = errors.New("there are no NetCluster Nodes in Env file")
		return
	}
	for _, node := range cluster.Nodes {
		err = node.CheckNetNode()
		if err != nil {
			return
		}
	}
	return
}

func (te *testEnv) CheckNetClusters() (err error) {
	if !(len(te.NetCluster) > 0) {
		err = errors.New("there are no NetCluster in Env file")
		return
	}
	for _, netCluster := range te.NetCluster {
		err = netCluster.CheckNetCluster()
		if err != nil {
			return
		}
	}
	return
}

func (te *testEnv) CheckCephConfiguration() (err error) {
	if te.CephConfiguration.ClusterName == "" ||
		te.CephConfiguration.NumberOfNodes < 3 ||
		te.CephConfiguration.PublicNetwork == "" ||
		te.CephConfiguration.ClusterNetwork == "" {
		err = errors.New("Error in configuration parameters (check: CephConfiguration.ClusterName, CephConfiguration.NumberOfNodes, CephConfiguration.PublicNetwork, CephConfiguration.ClusterNetwork)")
		return
	}
	return
}

func (te *testEnv) CheckPortusConfiguration() (err error) {
	if te.PortusConfiguration.FullyQualifiedDomainName == "" ||
		te.PortusConfiguration.Password == "" ||
		te.PortusConfiguration.SecretKeyBase == "" ||
		te.PortusConfiguration.AdminState == "" {
		err = errors.New("Error in configuration parameters (check: PortusConfiguration.FullyQualifiedDomainName " +
			"PortusConfiguration.Password " +
			"PortusConfiguration.SecretKeyBase " +
			"PortusConfiguration.AdminState)")
		return
	}
	return
}

func (te *testEnv) CheckAuthenticationProfile() (err error) {
	if te.AuthenticationProfile.Name == "" ||
		te.AuthenticationProfile.Type == "" {
		err = errors.New("Error in configuration parameters (check: AuthenticationProfile.Type, AuthenticationProfile.Name, CephConfiguration ClusterNetwork)")
		return
	}
	err = CheckLDAPConfiguration(&te.AuthenticationProfile.Profile)
	return
}

func CheckLDAPConfiguration(profile *pcc.LDAPConfiguration) (err error) {
	if profile.Domain == "" ||
		profile.UserIDAttribute == "" ||
		profile.UserBaseDN == "" ||
		profile.AnonymousBind == nil ||
		profile.BindDN == "" ||
		profile.BindPassword == "" ||
		profile.EncryptionPolicy == "" {
		err = errors.New("Error in configuration parameters (check: " +
			"Profile.UserIDAttribute " +
			"Profile.Domain " +
			"Profile.UserBaseDN " +
			"Profile.AnonymousBind" +
			"Profile.BindDN " +
			"Profile.BindPassword " +
			"Profile.EncryptionPolicy)")
		return
	}
	return
}

func (te *testEnv) CheckK8sAppConfiguration() (err error) {
	for _, app := range te.K8sAppConfiguration.Apps {
		err = CheckApp(&app)
		if err != nil {
			return
		}
	}
	return
}

func CheckApp(app *pcc.ConfigKApp) (err error) {
	if app.AppName == " " ||
		app.AppNamespace == " " ||
		app.GitUrl == " " ||
		app.GitRepoPath == " " ||
		app.GitBranch == " " ||
		app.Label == " " {
		err = errors.New("Error in configuration parameters (check: " +
			"app.AppName " +
			"app.AppNamespace " +
			"app.GitUrl " +
			"app.GitRepoPath" +
			"app.GitBranch " +
			"app.Label)")
		return
	}
	return
}

package capa

type ClusterConfig struct {
	ClusterDescription string         `json:"clusterDescription,omitempty"`
	ClusterName        string         `json:"clusterName,omitempty"`
	Organization       string         `json:"organization,omitempty"`
	AWS                *AWS           `json:"aws,omitempty"`
	Network            *Network       `json:"network,omitempty"`
	Bastion            *Bastion       `json:"bastion,omitempty"`
	ControlPlane       *ControlPlane  `json:"controlPlane,omitempty"`
	MachinePools       *[]MachinePool `json:"machinePools,omitempty"`
	FlatcarAWSAccount  string         `json:"flatcarAWSAccount,omitempty"`
	Proxy              *Proxy         `json:"proxy,omitempty"`
}

type DefaultAppsConfig struct {
	ClusterName  string `json:"clusterName,omitempty"`
	Organization string `json:"organization,omitempty"`
}

type AWS struct {
	Region string `json:"region,omitempty"`
	Role   string `json:"awsClusterRole,omitempty"`
}

type Network struct {
	AvailabilityZoneUsageLimit int      `json:"availabilityZoneUsageLimit,omitempty"`
	VPCCIDR                    string   `json:"vpcCIDR,omitempty"`
	TopologyMode               string   `json:"topologyMode,omitempty"`
	VPCMode                    string   `json:"vpcMode,omitempty"`
	ApiMode                    string   `json:"apiMode,omitempty"`
	DnsMode                    string   `json:"dnsMode,omitempty"`
	Subnets                    []string `json:"subnets,omitempty"`
}

type Bastion struct {
	InstanceType string `json:"instanceType,omitempty"`
	Replicas     int    `json:"replicas,omitempty"`
}

type ControlPlane struct {
	InstanceType           string   `json:"instanceType,omitempty"`
	Replicas               int      `json:"replicas,omitempty"`
	RootVolumeSizeGB       int      `json:"rootVolumeSizeGB,omitempty"`
	EtcdVolumeSizeGB       int      `json:"etcdVolumeSizeGB,omitempty"`
	ContainerdVolumeSizeGB int      `json:"containerdVolumeSizeGB,omitempty"`
	KubeletVolumeSizeGB    int      `json:"kubeletVolumeSizeGB,omitempty"`
	AvailabilityZones      []string `json:"availabilityZones,omitempty"`
}

type MachinePool struct {
	Name              string   `json:"name,omitempty"`
	AvailabilityZones []string `json:"availabilityZones,omitempty"`
	InstanceType      string   `json:"instanceType,omitempty"`
	MinSize           int      `json:"minSize,omitempty"`
	MaxSize           int      `json:"maxSize,omitempty"`
	RootVolumeSizeGB  int      `json:"rootVolumeSizeGB,omitempty"`
	CustomNodeLabels  []string `json:"customNodeLabels,omitempty"`
}

type Proxy struct {
	Enabled    bool   `json:"enabled,omitempty"`
	HttpsProxy string `json:"https_proxy,omitempty"`
	HttpProxy  string `json:"http_proxy,omitempty"`
	NoProxy    string `json:"no_proxy,omitempty"`
}

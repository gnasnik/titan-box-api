package model

import "time"

type Box struct {
	Username          string      `json:"-" db:"username"`
	BoxId             string      `json:"boxId" db:"boxId"`
	SupplierBoxId     string      `json:"supplierBoxId" db:"supplierBoxId"`
	Online            string      `json:"online" db:"online"`
	TcpNatType        string      `json:"tcpNatType" db:"tcpNatType"`
	UdpNatType        string      `json:"udpNatType" db:"udpNatType"`
	PublicIp          string      `json:"publicIp" db:"publicIp"`
	PrivateIp         string      `json:"privateIp" db:"privateIp"`
	Isp               string      `json:"isp" db:"isp"`
	Province          string      `json:"province" db:"province"`
	City              string      `json:"city" db:"city"`
	CpuArch           string      `json:"cpuArch" db:"cpuArch"`
	CpuCores          string      `json:"cpuCores" db:"cpuCores"`
	MemorySize        string      `json:"memorySize" db:"memorySize"`
	DiskInfos         []*DiskInfo `json:"diskInfos" db:"-"`
	Os                string      `json:"os" db:"os"`
	PluginVersion     string      `json:"pluginVersion" db:"pluginVersion"`
	PluginDeployTime  string      `json:"pluginDeployTime" db:"pluginDeployTime"`
	ProcessStatus     string      `json:"processStatus" db:"processStatus"`
	PlanTask          string      `json:"planTask" db:"planTask"`
	PressBandwidth    float64     `json:"pressBandwidth" db:"pressBandwidth"`
	Fault             string      `json:"fault" db:"fault"`
	Upload            float64     `json:"upload" db:"upload"`
	Download          float64     `json:"download" db:"download"`
	DiskUsage         float64     `json:"diskUsage" db:"diskUsage"`
	Upnp              bool        `json:"upnp" db:"upnp"`
	NotDeployReason   string      `json:"notDeployReason" db:"notDeployReason"`
	ReportUpBandwidth float64     `json:"reportUpBandwidth" db:"reportUpBandwidth"`
	Remark            string      `json:"remark" db:"remark"`
	Icmpv6Out         float64     `json:"icmpv6Out" db:"icmpv6Out"`
	CreatedAt         time.Time   `json:"-" db:"createdAt"`
	UpdatedAt         time.Time   `json:"-" db:"updatedAt"`
}

type DiskInfo struct {
	BoxId         string `json:"-" db:"boxId"`
	SupplierBoxId string `json:"-"  db:"supplierBoxId"`
	DiskId        string `json:"diskId" db:"diskId"`
	DiskSize      string `json:"diskSize" db:"diskSize"`
	DiskMedia     string `json:"diskMedia" db:"diskMedia"`
	DiskUsed      string `json:"diskUsed" db:"diskUsed"`
}

type BoxIncome struct {
	Username       string    `json:"-" db:"username"`
	Date           string    `json:"date" db:"date"`
	BoxId          string    `json:"boxId"  db:"boxId"`
	Remark         string    `json:"remark" db:"remark"`
	Bw             string    `json:"bw" db:"bw"`
	Amount         string    `json:"amount" db:"amount"`
	SupplierBoxId  string    `json:"supplierBoxId" db:"supplierBoxId"`
	BwAmount       string    `json:"bwAmount" db:"bwAmount"`
	ActivityIncome string    `json:"activityIncome" db:"activityIncome"`
	UserRemark     string    `json:"userRemark" db:"userRemark"`
	DistAmount     string    `json:"distAmount" db:"distAmount"`
	DistPercent    int       `json:"distPercent" db:"distPercent"`
	InviterId      string    `json:"inviterId" db:"inviterId"`
	UpdatedAt      time.Time `json:"-" db:"updatedAt"`
}

type BoxBandwidth struct {
	Username      string    `json:"-" db:"username"`
	BoxId         string    `json:"-" db:"boxId"`
	SupplierBoxId string    `json:"-" db:"supplierBoxId"`
	Time          string    `json:"time" db:"time"`
	Upload        float64   `json:"upload" db:"upload"`
	Download      float64   `json:"download" db:"download"`
	UpdatedAt     time.Time `json:"-" db:"updatedAt"`
}

type BoxQuality struct {
	Username      string    `json:"-" db:"username"`
	BoxId         string    `json:"-" db:"boxId"`
	SupplierBoxId string    `json:"-" db:"supplierBoxId"`
	Time          string    `json:"time" db:"time"`
	PacketLoss    float64   `json:"packetLoss" db:"packetLoss"`
	TcpNatType    string    `json:"tcpNatType" db:"tcpNatType"`
	UdpNatType    string    `json:"udpNatType" db:"udpNatType"`
	CpuUsage      float64   `json:"cpuUsage" db:"cpuUsage"`
	MemoryUsage   float64   `json:"memoryUsage" db:"memoryUsage"`
	DiskUsage     float64   `json:"diskUsage" db:"diskUsage"`
	UpdatedAt     time.Time `json:"-" db:"updatedAt"`
}

package network

type Peer struct {
	Name    string `json:"Name"`
	Id      string `json:"Id"`
	Ip      string `json:"Ip"`
	LocalIp string `json:"LocalIp"`
	State   string `json:"State"`
	Port    int    `json:"Port"`
}
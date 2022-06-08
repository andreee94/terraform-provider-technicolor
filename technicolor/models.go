package technicolor

type PortForwarded struct {
	Enabled  bool
	Name     string
	Protocol string
	WanPort  int
	LanPort  int
	LanIp    string
	LanMac   string
}

type PortForwardedWithIndex struct {
	Index int
	Data  PortForwarded
}

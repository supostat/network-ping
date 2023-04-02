package network_ping

import (
	"fmt"
	"net"
	"sync"
	"time"

	utils "github.com/supostat/network-ping/utils"
)

var (
	protocols = []string{"tcp"}
	timeout   = time.Microsecond * 50000
	wg        sync.WaitGroup
)

type NetworkPing struct {
	CIDRs               []string
	ports               []int
	OnConnect           func(*Result)
	OnConnectionRefused func(*Result)
}

type Result struct {
	ipAddr   string
	protocol string
	port     int
	status   string
}

func New(ipParams string, ports []int) *NetworkPing {
	CIDRs, _ := utils.GetCIDRs(ipParams)

	return &NetworkPing{
		CIDRs: CIDRs,
		ports: ports,
	}
}

func (np *NetworkPing) Start() {
	var ip net.IP
	var ipNet *net.IPNet

	var incIP = func(ip net.IP) {
		for j := len(ip) - 1; j >= 0; j-- {
			ip[j]++
			if ip[j] > 0 {
				break
			}
		}
	}
	for _, cidr := range np.CIDRs {
		ip, ipNet, _ = net.ParseCIDR(cidr)

		for ip := ip.Mask(ipNet.Mask); ipNet.Contains(ip); incIP(ip) {
			wg.Add(1)
			go func(ip string) {
				defer wg.Done()

				for _, protocol := range protocols {
					for _, port := range np.ports {
						addr := fmt.Sprintf("%s:%d", ip, port)

						c, e := net.DialTimeout(protocol, addr, timeout)
						if e != nil {
							handler := np.OnConnectionRefused
							if handler != nil {
								handler(&Result{
									ipAddr:   ip,
									protocol: protocol,
									port:     port,
									status:   "refused",
								})
							}
						}
						if e == nil {
							handler := np.OnConnect
							if handler != nil {
								handler(&Result{
									ipAddr:   ip,
									protocol: protocol,
									port:     port,
									status:   "opened",
								})
							}
							c.Close()
						}
					}
				}
			}(ip.String())
		}

		wg.Wait()

	}

}

func ParseIP(s string) string {
	i := net.ParseIP(s)
	return i.String()
}

func ParseCIDR(s string) (string, error) {
	_, ipv4Net, err := net.ParseCIDR(s)
	if err != nil {
		return "", err
	}
	return ipv4Net.String(), nil
}

func ParseAddress(addr string) (string, bool) {
	cidr, err := ParseCIDR(addr)
	if err == nil {
		return cidr, true
	}

	ip := ParseIP(addr)
	if ip != "" {
		return ip, true
	}

	return "", false
}

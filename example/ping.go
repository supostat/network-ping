package main

import (
	"log"

	network_ping "github.com/supostat/network-ping"
)

func main() {
	ports := []int{80, 443}
	np := network_ping.New("192.168.100.1-192.168.100.255", ports)

	np.OnConnect = func(res *network_ping.Result) {
		log.Println(res)
	}
	np.OnConnectionRefused = func(res *network_ping.Result) {
		log.Println(res)
	}
	np.Start()
}

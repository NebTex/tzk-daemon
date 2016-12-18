package main

import "strings"
import "net"

//GetLocalAddresses  add the ipv4/ipv6 address associates to the current
//interface
func (f *Facts) GetLocalAddresses() {
	//var ra []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
	}
	for _, addr := range addrs {
		f.AddAddress(strings.Split(addr.String(), "/")[0])
	}
}

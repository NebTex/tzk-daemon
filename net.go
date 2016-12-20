package main

import "strings"
import "net"

//GetLocalAddresses  add the ipv4/ipv6 address associates to the current
//interface
func (f *Facts) GetLocalAddresses() {
	//var ra []string
	addresses, err := net.InterfaceAddrs()
	checkFatal(err)

	for _, address := range addresses {
		f.AddAddress(strings.Split(address.String(), "/")[0])
	}
}

package main

import (
	"fmt"
	"math/rand"
	"net"
	"runtime"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/consul/api"
)

//DefaultSubnet the default cidr of the vpn
const DefaultSubnet = "10.187.0.0/16"

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func cidrAddress(cidr string, TakenAddresses []string) []string {
	ips := []string{}

	ip, ipNet, err := net.ParseCIDR(cidr)

	if err != nil {
		log.Fatal(err)
	}

	for ip = ip.Mask(ipNet.Mask); ipNet.Contains(ip); inc(ip) {
		if !StringSliceContains(TakenAddresses, ip.String()) {
			ips = append(ips, ip.String())
		}
	}
	return ips
}

func currentIP(client *api.Client, c Config, hostname string) *string {
	var ret string
	// check if hostname has a ip associate with it
	kv, _, err := client.KV().
		Get(fmt.Sprintf("%s/Hosts/%s/VpnAddress", c.Vpn.Name, hostname), nil)
	if err != nil {
		log.Fatal(err)
	}
	if kv != nil {
		ret = string(kv.Value)
		return &ret
	}
	return nil
}
func checkFatal(e error) {
	_, file, line, _ := runtime.Caller(1)
	if e != nil {
		log.WithFields(log.Fields{"file": file,
			"line": line,
		}).Fatal(e)
	}

}

func assignIP(client *api.Client, c Config, hostname string) string {
	var newIP string

	ci := currentIP(client, c, hostname)
	if ci != nil {
		//check ip is register in TakenAddresses prefix
		kv, _, err := client.KV().
			Get(fmt.Sprintf("%s/TakenAddresses/%s", c.Vpn.Name, *ci), nil)
		checkFatal(err)
		if kv != nil {
			if string(kv.Value) == hostname {
				return *ci
				/*sKv, _, err := client.KV().
					Get(fmt.Sprintf("%s/Subnet", c.Vpn.Name), nil)

				checkFatal(err)

				subnet := string(sKv.Value)
				_, ipNet, err := net.ParseCIDR(subnet)
				checkFatal(err)

				if ipNet.Contains(net.ParseIP(*ci)) {
					log.Info(fmt.Sprintf("%s is associated to this host", string(kv.Value)))
					return *ci
				}

				// delete taken ip
				_, err = client.KV().
					Delete(fmt.Sprintf("%s/TakenAddresses/%s", c.Vpn.Name, *ci), nil)
				checkFatal(err)*/

			}
		}

	}
	// return subnet
	kv, _, err := client.KV().Get(fmt.Sprintf("%s/Subnet", c.Vpn.Name), nil)
	checkFatal(err)
	subnet := string(kv.Value)

	// return all the allocated addresses
	kvs, _, err := client.KV().
		List(fmt.Sprintf("%s/TakenAddresses", c.Vpn.Name), nil)

	if err != nil {
		log.Fatal(err)
	}
	return takeAddress(kvs, subnet, newIP, client, c, hostname)
}

func takeAddress(kvs api.KVPairs, subnet string,
	newIP string, client *api.Client, c Config, hostname string) string {
	rand.Seed(time.Now().UnixNano())

	takenAddresses := []string{}
	for _, kv := range kvs {
		takenAddresses = append(takenAddresses, kv.Key)
	}

	for {
		availableIPs := cidrAddress(subnet, takenAddresses)
		if len(availableIPs) == 0 {
			log.Fatal("Not available address to take")
		}
		newIP = availableIPs[rand.Intn(len(availableIPs))]

		kv := &api.KVPair{}
		kv.CreateIndex = 0
		kv.Key = fmt.Sprintf("%s/TakenAddresses/%s", c.Vpn.Name, newIP)
		kv.Value = []byte(hostname)

		//save new ip and lock it
		ok, _, err := client.KV().CAS(kv, nil)
		checkFatal(err)

		if ok {
			return newIP
		}
		takenAddresses = append(takenAddresses, newIP)
		time.Sleep(100 * time.Millisecond)
		log.Info("Requesting new IP")
	}
}

//DHCP pick a ip address from the subnet cidr
//use the locking features of consul to guarantee
//ip per node
func DHCP(c Config, hostname string) string {
	client := getConsulClient(c)

	address := assignIP(client, c, hostname)
	//store ip address in the host prefix
	kv := &api.KVPair{}
	kv.Key = fmt.Sprintf("%s/Hosts/%s/VpnAddress", c.Vpn.Name, hostname)
	kv.Value = []byte(address)
	_, err := client.KV().Put(kv, nil)
	checkFatal(err)

	log.Infof("The ip address %s have been assigned to this node", address)
	return address

}

func initSubnet(c Config) {
	client := getConsulClient(c)
	kv, _, err := client.KV().Get(fmt.Sprintf("%s/Subnet", c.Vpn.Name), nil)
	checkFatal(err)
	s := DefaultSubnet
	if c.Vpn.Subnet != "" {
		s = c.Vpn.Subnet
	}
	if kv == nil {
		kv = &api.KVPair{Key: fmt.Sprintf("%s/Subnet", c.Vpn.Name),
			Value: []byte(s)}
		_, err = client.KV().Put(kv, nil)
		checkFatal(err)

	}

}

package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/criloz/goblin"
	"github.com/hashicorp/consul/api"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"net"
	"strings"
	"testing"
	"time"
)

func bootstrapConsul(subnet string, vpnName string, c Config) {
	client := getConsulClient(c)

	_, err := client.KV().DeleteTree(vpnName, nil)
	if err != nil {
		log.Fatal(err)
	}
	kp := &api.KVPair{Key: fmt.Sprintf("%s/Subnet", vpnName),
		Value: []byte(subnet)}

	_, err = client.KV().Put(kp, nil)
	if err != nil {
		log.Fatal(err)
	}

	f := Facts{}
	f.GetContainerStatus()
	f.GetLocalAddresses()
	f.GetGeoIP()
	f.PublicKey = "pubkey"
	f.Hostname = "node1"
	f.SendToConsul(c)
	h := Host{Facts: f}
	h.SetConfigConsul(c)

}
func checkFail(g *goblin.G, err error) {
	if err != nil {
		g.Fail(err)
	}
}

//TestDHCP
func TestDHCP(t *testing.T) {
	g := goblin.Goblin(t)
	c := Config{}
	c.Consul.Address = "localhost:8500"
	c.Consul.Scheme = "http"
	c.Vpn.Name = "tzn"
	client := getConsulClient(c)

	g.Describe("DHCP", func() {
		g.It("Should pick an ip from the subnet", func() {
			subnets := []string{
				"10.65.0.0/16",
				"169.247.0.0/16",
				"172.120.4.0/28",
				"10.20.2.0/24"}
			for _, subnet := range subnets {
				bootstrapConsul(subnet, c.Vpn.Name, c)
				hostname := "node1"
				DHCP(c, hostname)
				hKey := fmt.Sprintf("%s/Hosts/%s/VpnAddress", c.Vpn.Name, hostname)
				hkv, _, err := client.KV().Get(hKey, nil)
				checkFail(g, err)
				vKey := fmt.Sprintf("%s/TakenAddresses/%s",
					c.Vpn.Name, string(hkv.Value))
				vkv, _, err := client.KV().Get(vKey, nil)
				checkFail(g, err)
				assert.Equal(g, string(vkv.Value), hostname)
				_, ipNet, err := net.ParseCIDR(subnet)
				checkFail(g, err)
				assert.True(g, ipNet.Contains(net.ParseIP(string(hkv.Value))))
			}
		})

		g.It("Should preserve the ip", func() {
			subnet := "10.65.1.0/24"
			bootstrapConsul(subnet, c.Vpn.Name, c)
			assert.Equal(g, DHCP(c, "node1"), DHCP(c, "node1"))

		})
		g.It("Should pick  all the ip from the subnet", func() {
			g.Timeout(60 * time.Second)

			subnet := "10.65.1.0/24"
			ch := make(chan int, 2)
			bootstrapConsul(subnet, c.Vpn.Name, c)
			for i := 0; i < 256; i++ {
				hostname := uuid.NewV4().String()
				go func() {
					DHCP(c, hostname)
					ch <- 1
				}()

			}
			total := 0

			for elem := range ch {
				total += elem
				if total >= 256 {
					close(ch)
					break
				}
			}
			// return all the allocated addresses
			kvs, _, err := client.KV().
				List(fmt.Sprintf("%s/TakenAddresses", c.Vpn.Name), nil)

			checkFail(g, err)

			assert.Equal(g, len(kvs), 256)
			_, ipNet, err := net.ParseCIDR(subnet)
			checkFail(g, err)

			for _, kv := range kvs {
				ss := strings.Split(kv.Key, "/")
				assert.True(g, ipNet.Contains(net.ParseIP(ss[len(ss)-1])))

			}
		})

	})

}

func TestDHCPChangeSubnet(t *testing.T) {
	g := goblin.Goblin(t)
	c := Config{}
	c.Consul.Address = "localhost:8500"
	c.Consul.Scheme = "http"
	c.Vpn.Name = "tzn"
	client := getConsulClient(c)

	g.Describe("DHCP", func() {
		g.It("Should change ip if the subnet changed", func() {
			g.Timeout(30 * time.Second)
			subnet1 := "10.65.1.0/24"
			subnet2 := "10.1.0.0/16"
			_, ipNet1, err := net.ParseCIDR(subnet1)
			checkFail(g, err)
			_, ipNet2, err := net.ParseCIDR(subnet2)
			checkFail(g, err)
			bootstrapConsul(subnet1, c.Vpn.Name, c)
			newIP1 := DHCP(c, "node1")
			assert.True(g, ipNet1.Contains(net.ParseIP(newIP1)))
			_, err = client.KV().
				Put(&api.KVPair{Key: "tzn/Subnet", Value: []byte(subnet2)}, nil)
			checkFail(g, err)
			DHCP(c, "node1")
			newIP2 := DHCP(c, "node1")
			assert.NotEqual(g, newIP2, newIP1)
			assert.True(g, ipNet2.Contains(net.ParseIP(newIP2)))
			kv, _, err := client.KV().
				Get(fmt.Sprintf("tzn/TakenAddresses/%s", newIP1), nil)
			checkFail(g, err)
			assert.Nil(g, kv)
			kv, _, err = client.KV().
				Get(fmt.Sprintf("tzn/TakenAddresses/%s", newIP2), nil)
			checkFail(g, err)
			assert.NotNil(g, kv)

		})

	})

}
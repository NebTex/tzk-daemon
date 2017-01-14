package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/criloz/goblin"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
)

func addHost(c Config, hostname string, pubkey string, addresses ...string) {
	h := Host{}
	h.Facts.GetContainerStatus()
	h.Facts.Addresses = make(map[string]string)
	h.Facts.HasChanged = true
	//h.Facts.GetGeoIP()
	h.Facts.PublicKey = pubkey
	h.Facts.Hostname = hostname
	for _, address := range addresses {
		h.Facts.Addresses[address] = ""
	}
	h.Facts.SendToConsul(c)
	h.SetConfigConsul(c)
	DHCP(c, h.Facts.Hostname)
}

//TestGenerateFiles
func TestGenerateFiles(t *testing.T) {
	g := goblin.Goblin(t)
	c := Config{}
	c.Consul.Address = "localhost:8500"
	c.Consul.Scheme = "http"
	c.Vpn.Name = "tzk"
    
	g.Describe("GenerateFiles", func() {
		g.It("Should generate the files need to run tinc", func(done goblin.Done) {
			g.Timeout(60 * time.Second)
			handle := func(v *Vpn, close func()) {
				files := v.GenerateFiles("node1", c)
				ip := DHCP(c, "node1")
				ip2 := DHCP(c, "node2")
				ip3 := DHCP(c, "node3")
				ip4 := DHCP(c, "node4")

				assert.Equal(g, fmt.Sprintf(`#!/bin/sh
ip link set $INTERFACE up
ip addr add  %s dev $INTERFACE
ip route add 10.1.0.0/16 dev $INTERFACE`, ip), files.Tinc["tinc-up"])

				assert.Equal(g, fmt.Sprintf(`#!/bin/sh
ip route del 10.1.0.0/16 dev $INTERFACE
ip addr del %s dev $INTERFACE
ip link set $INTERFACE down`, ip), files.Tinc["tinc-down"])

				assert.Equal(g, `AddressFamily=ipv4
AutoConnect=yes
Device=/dev/net/tun
LocalDiscovery=yes
MaxTimeout=300
Mode=router
Name=node1`, files.Tinc["tinc.conf"])

				assert.Equal(g, fmt.Sprintf(`Ed25519PublicKey=pubkey2
Address=185.36.25.14
Subnet=%s`, ip2), files.Hosts["node2"])
				assert.Equal(g, fmt.Sprintf(`Ed25519PublicKey=pubkey3
Address=108.36.25.14
Subnet=%s`, ip3), files.Hosts["node3"])
				assert.Equal(g, fmt.Sprintf(`Ed25519PublicKey=pubkey4
Address=95.36.25.14
Subnet=%s`, ip4), files.Hosts["node4"])
				close()
				done()
			}
			bootstrapConsul("10.1.0.0/16", c.Vpn.Name, c)
			DHCP(c, "node1")
			addHost(c, "node2", "pubkey2", "185.36.25.14")
			addHost(c, "node3", "pubkey3", "108.36.25.14")
			addHost(c, "node4", "pubkey4", "95.36.25.14")
			go WatchConsul(c, handle)
		})
	})
}

//TestCompareFiles
func TestCompareFiles(t *testing.T) {
	g := goblin.Goblin(t)
	c := Config{}
	c.Consul.Address = "localhost:8500"
	c.Consul.Scheme = "http"
	c.Vpn.Name = "tzk"

	client := getConsulClient(c)

	g.Describe("CompareFiles", func() {
		g.It("Should return true with the same inputs ", func(done goblin.Done) {
			g.Timeout(15 * time.Second)
			handle := func(v *Vpn, close func()) {
				files := v.GenerateFiles("node1", c)
				assert.True(g, files.Equal(files))
				assert.True(g, files.Equal(files))
				assert.True(g, files.Equal(files))
				assert.True(g, files.Equal(files))
				close()
				done()

			}
			bootstrapConsul("10.1.0.0/16", c.Vpn.Name, c)
			DHCP(c, "node1")
			addHost(c, "node2", "pubkey2", "185.36.25.14")
			addHost(c, "node3", "pubkey3", "108.36.25.14")
			addHost(c, "node4", "pubkey4", "95.36.25.14")
			go WatchConsul(c, handle)
		})
		g.It("should detect changes ", func(done goblin.Done) {
			g.Timeout(10 * time.Second)
			count := 0
			var files1, files2 *Files
			handle := func(v *Vpn, close func()) {
				count++
				files := v.GenerateFiles("node1", c)
				if count == 1 {
					assert.True(g, files.Equal(files))
					files1 = files
				}
				if count == 2 {

					assert.False(g, files.Equal(files1))
					files2 = files
				}

				if count == 3 {
					assert.False(g, files.Equal(files2))
					close()
					done()
				}
			}
			bootstrapConsul("10.1.0.0/16", c.Vpn.Name, c)
			DHCP(c, "node1")
			addHost(c, "node2", "pubkey2", "185.36.25.14")
			addHost(c, "node3", "pubkey3", "108.36.25.14")
			addHost(c, "node4", "pubkey4", "95.36.25.14")
			go WatchConsul(c, handle)
			time.Sleep(1 * time.Second)
			_, err := client.KV().
				Put(&api.KVPair{Key: "tzk/Hosts/node1/Facts/PublicKey",
					Value: []byte("xxxxxxxx")}, nil)
			checkFail(g, err)
			time.Sleep(1 * time.Second)
			addHost(c, "node5", "pubkey5", "95.36.25.14")

		})

	})
}

//TestSaveMap
func TestSaveMap(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("saveMap", func() {
		g.It("Should save all the map elements ", func() {
			m := map[string]string{"1.test": "ok1", "2.test": "ok2",
				"3.test": "ok3"}
			err := saveMap(m, "/tmp")
			if err != nil {
				g.Fail(err)
			}
			for k, v := range m {
				d, err := ioutil.ReadFile(filepath.Join("/tmp", k))
				if err != nil {
					g.Fail(err)
				}
				assert.Equal(g, v, string(d))
			}

		})

	})
}

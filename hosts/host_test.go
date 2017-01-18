package hosts

import (
    "fmt"
    "github.com/criloz/goblin"
    "github.com/stretchr/testify/assert"
    "strings"
    "testing"
    "tzk-daemon/commons"
    "tzk-daemon/dhcp"
)

func addHost(c commons.Config, hostname string, pubkey string,
addresses ...string) {
    h := commons.Host{}
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
    dhcp.DHCP(c, h.Facts.Hostname)
}

//TestParseToFileFormat
func TestParseToFileFormat(t *testing.T) {
    g := goblin.Goblin(t)
    c := commons.Config{}
    c.Consul.Address = "localhost:8500"
    c.Consul.Scheme = "http"
    c.Vpn.Name = "tzk"
    g.Describe("ParseToFileFormat", func() {
        g.It("should insert all the nodes in the host file",
            func(done goblin.Done) {
                handle := func(v *Vpn, close func()) {
                    o := v.Hosts.parseToFileFormat(v.Hosts["node1"])
                    assert.True(g, strings.Contains(o, "node3.tzk.local"))
                    assert.True(g, strings.Contains(o, "node4.tzk.local"))
                    assert.True(g, strings.Contains(o, "node1.tzk.local"))
                    assert.True(g, strings.Contains(o, "node2.tzk.local"))
                    assert.True(g, strings.Contains(o,
                        fmt.Sprintf("%s\tnode1", v.Hosts["node1"].VpnAddress)))
                    
                    close()
                    done()
                }
                subnet1 := "10.65.1.0/24"
                c.Vpn.Subnet = subnet1
                commons.BootstrapConsul(c.Vpn.Name, c)
                dhcp.DHCP(c, "node1")
                addHost(c, "node2", "pubkey2", "185.36.25.14")
                addHost(c, "node3", "pubkey3", "108.36.25.14")
                addHost(c, "node4", "pubkey4", "95.36.25.14")
                go WatchConsul(c, handle)
                
            })
        
    })
}

//TestManageHostBlock
func TestManageHostBlock(t *testing.T) {
    g := goblin.Goblin(t)
    c := commons.Config{}
    c.Consul.Address = "localhost:8500"
    c.Consul.Scheme = "http"
    c.Vpn.Name = "tzk"
    g.Describe("ParseToFileFormat", func() {
        g.It("should insert or remplace the block in the host file",
            func(done goblin.Done) {
                handle := func(v *Vpn, close func()) {
                    o := v.Hosts.manageHostBlock("", v.Hosts["node 1"])
                    assert.True(g, strings.Contains(o, "node3.tzk.local"))
                    assert.True(g, strings.Contains(o, "node4.tzk.local"))
                    assert.True(g, strings.Contains(o, "node1.tzk.local"))
                    assert.True(g, strings.Contains(o, "node2.tzk.local"))
                    
                    o = v.Hosts.manageHostBlock(`127.0.0.1 localhost
#/tzk/NoEdit
                    10.65.1.23	node20.tzk.local
                    10.65.1.206	node30.tzk.local
                    10.65.1.248	node40.tzk.local
                    #/tzk/NoEdit`, v.Hosts["node 1"])
                    assert.True(g, strings.Contains(o, "node3.tzk.local"))
                    assert.True(g, strings.Contains(o, "node4.tzk.local"))
                    assert.True(g, strings.Contains(o, "node1.tzk.local"))
                    assert.True(g, strings.Contains(o, "node2.tzk.local"))
                    assert.False(g, strings.Contains(o, "node20.tzk.local"))
                    assert.False(g, strings.Contains(o, "node30.tzk.local"))
                    assert.False(g, strings.Contains(o, "node40.tzk.local"))
                    
                    close()
                    done()
                    
                }
                c.Vpn.Subnet = "10.65.1.0/24"
                commons.BootstrapConsul(c.Vpn.Name, c)
                dhcp.DHCP(c, "node1")
                addHost(c, "node2", "pubkey2", "185.36.25.14")
                addHost(c, "node3", "pubkey3", "108.36.25.14")
                addHost(c, "node4", "pubkey4", "95.36.25.14")
                go WatchConsul(c, handle)
                
            })
        
    })
}

//TestFixName
func TestFixName(t *testing.T) {
    g := goblin.Goblin(t)
    g.Describe("fixName", func() {
        g.It("should replace non standard characters by _",
            func() {
                assert.Equal(g, FixName("ikkk?ikjkj-ljl{{"), "ikkk_ikjkj_ljl__")
            })
        
    })
}

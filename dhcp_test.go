package main

import (
    "fmt"
    "io/ioutil"
    "net"
    "strings"
    "testing"
    "time"
    
    log "github.com/Sirupsen/logrus"
    "github.com/criloz/goblin"
    "github.com/hashicorp/consul/api"
    "github.com/satori/go.uuid"
    "github.com/stretchr/testify/assert"
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
    c.Vpn.PublicKeyFile = "/tmp/pubkey"
    err = ioutil.WriteFile(c.Vpn.PublicKeyFile, []byte("pubkey = pubkey"), 0755)
    
    checkFatal(err)
    
    f := Facts{}
    f.GetContainerStatus()
    f.GetLocalAddresses()
    f.GetGeoIP()
    f.GetTincInfo(c, func() (string, error) {
        return "node1", nil
    })
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
    c.Vpn.Name = "tzk"
    client := getConsulClient(c)
    
    g.Describe("DHCP", func() {
        g.It("Should pick an ip from the subnet", func() {
            g.Timeout(60 * time.Second)
            subnets := []string{
                "10.65.0.0/16",
                "169.247.0.0/16",
                "172.120.4.0/28",
                "10.20.2.0/24"}
            for _, subnet := range subnets {
                bootstrapConsul(subnet, c.Vpn.Name, c)
                hostname := "node1"
                DHCP(c, hostname)
                hKey := fmt.Sprintf("%s/Hosts/%s/VpnAddress", c.Vpn.Name,
                    hostname)
                hkv, _, err := client.KV().Get(hKey, nil)
                checkFail(g, err)
                
                assert.NotNil(g, hkv)
                
                vKey := fmt.Sprintf("%s/TakenAddresses/%s",
                    c.Vpn.Name, string(hkv.Value))
                vkv, _, err := client.KV().Get(vKey, nil)
                checkFail(g, err)
                assert.NotNil(g, vkv)
                assert.Equal(g, string(vkv.Value), hostname)
                _, ipNet, err := net.ParseCIDR(subnet)
                checkFail(g, err)
                assert.True(g, ipNet.Contains(net.ParseIP(string(hkv.Value))))
            }
        })
        
        g.It("Should preserve the ip", func() {
            g.Timeout(60 * time.Second)
            subnet := "10.65.1.0/24"
            bootstrapConsul(subnet, c.Vpn.Name, c)
            assert.Equal(g, DHCP(c, "node1"), DHCP(c, "node1"))
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
                assert.True(g, ipNet.Contains(net.ParseIP(ss[len(ss) - 1])))
                
            }
        })
        
    })
    
}

func TestDHCPChangeSubnet(t *testing.T) {
    g := goblin.Goblin(t)
    c := Config{}
    c.Consul.Address = "localhost:8500"
    c.Consul.Scheme = "http"
    c.Vpn.Name = "tzk"
    client := getConsulClient(c)
    
    g.Describe("DHCP", func() {
        g.It("Should not change ip if the subnet changed", func() {
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
                Put(&api.KVPair{Key: "tzk/Subnet", Value: []byte(subnet2)}, nil)
            checkFail(g, err)
            DHCP(c, "node1")
            newIP2 := DHCP(c, "node1")
            assert.Equal(g, newIP2, newIP1)
            assert.False(g, ipNet2.Contains(net.ParseIP(newIP2)))
            kv, _, err := client.KV().
                Get(fmt.Sprintf("tzk/TakenAddresses/%s", newIP1), nil)
            checkFail(g, err)
            assert.NotNil(g, kv)
        })
    })
}

func TestAssignStaticIP(t *testing.T) {
    g := goblin.Goblin(t)
    c := Config{}
    c.Consul.Address = "localhost:8500"
    c.Consul.Scheme = "http"
    c.Vpn.Name = "tzk"
    c.Vpn.NodeIP = "10.187.0.50"
    client := getConsulClient(c)
    _, err := client.KV().DeleteTree(c.Vpn.Name, nil)
    checkFatal(err)
    bootstrapConsul(DefaultSubnet, c.Vpn.Name, c)
    
    g.Describe("initSubnet", func() {
        g.It("Should pick the NodeIP", func() {
            hostname := "node1"
            DHCP(c, hostname)
            hKey := fmt.Sprintf("%s/Hosts/%s/VpnAddress", c.Vpn.Name, hostname)
            hkv, _, err := client.KV().Get(hKey, nil)
            checkFail(g, err)
            assert.Equal(g, string(hkv.Value), c.Vpn.NodeIP)
        })
        g.It("Should pick another ip if is taken", func() {
            hostname := "node2"
            DHCP(c, hostname)
            hKey := fmt.Sprintf("%s/Hosts/%s/VpnAddress", c.Vpn.Name, hostname)
            hkv, _, err := client.KV().Get(hKey, nil)
            checkFail(g, err)
            assert.NotEqual(g, string(hkv.Value), c.Vpn.NodeIP)
        })
    
        g.It("Should not use the ip if is not valid", func() {
            hostname := "node3"
            c.Vpn.NodeIP = "test"
            DHCP(c, hostname)
            hKey := fmt.Sprintf("%s/Hosts/%s/VpnAddress", c.Vpn.Name, hostname)
            hkv, _, err := client.KV().Get(hKey, nil)
            checkFail(g, err)
            assert.NotEqual(g, string(hkv.Value), c.Vpn.NodeIP)
        })
    
        g.It("Should not use the ip if is in the current subnet", func() {
            hostname := "node5"
            c.Vpn.NodeIP = "10.32.3.3"
            DHCP(c, hostname)
            hKey := fmt.Sprintf("%s/Hosts/%s/VpnAddress", c.Vpn.Name, hostname)
            hkv, _, err := client.KV().Get(hKey, nil)
            checkFail(g, err)
            assert.NotEqual(g, string(hkv.Value), c.Vpn.NodeIP)
        })
        
    })
    
}

func TestInitSubnet(t *testing.T) {
    g := goblin.Goblin(t)
    c := Config{}
    c.Consul.Address = "localhost:8500"
    c.Consul.Scheme = "http"
    c.Vpn.Name = "tzk"
    client := getConsulClient(c)
    
    g.Describe("initSubnet", func() {
        g.It("Should assign the default subnet", func() {
            _, err := client.KV().DeleteTree(c.Vpn.Name, nil)
            if err != nil {
                log.Fatal(err)
            }
            initSubnet(c)
            kv, _, err := client.KV().Get(fmt.Sprintf("%s/Subnet", c.Vpn.Name),
                nil)
            checkFail(g, err)
            assert.NotNil(g, kv)
            assert.Equal(g, string(kv.Value), DefaultSubnet)
        })
        
        g.It("Should not change subnet if already exist", func() {
            _, err := client.KV().DeleteTree(c.Vpn.Name, nil)
            if err != nil {
                log.Fatal(err)
            }
            kv := &api.KVPair{Key: fmt.Sprintf("%s/Subnet", c.Vpn.Name),
                Value: []byte("10.1.0.0/16")}
            _, err = client.KV().Put(kv, nil)
            checkFail(g, err)
            initSubnet(c)
            kv, _, err = client.KV().Get(fmt.Sprintf("%s/Subnet", c.Vpn.Name),
                nil)
            checkFail(g, err)
            assert.NotEqual(g, string(kv.Value), DefaultSubnet)
        })
        
    })
    
}

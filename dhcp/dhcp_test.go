package dhcp

import (
    "fmt"
    "net"
    "strings"
    "testing"
    "time"
    "tzk-daemon/commons"
    log "github.com/Sirupsen/logrus"
    "github.com/criloz/goblin"
    "github.com/hashicorp/consul/api"
    "github.com/satori/go.uuid"
    "github.com/stretchr/testify/assert"
)



//TestDHCP
func TestDHCP(t *testing.T) {
    g := goblin.Goblin(t)
    c := commons.Config{}
    c.Consul.Address = "localhost:8500"
    c.Consul.Scheme = "http"
    c.Vpn.Name = "tzk"
    client := commons.GetConsulClient(c)
    
    g.Describe("DHCP", func() {
        g.It("Should pick an ip from the subnet", func() {
            g.Timeout(60 * time.Second)
            subnets := []string{
                "10.65.0.0/16",
                "169.247.0.0/16",
                "172.120.4.0/28",
                "10.20.2.0/24"}
            
            ClusterCIDR := []string{
                "10.65.0.0/16",
                "10.32.0.0/16",
                "172.17.0.0/12",
                "10.20.0.0/16"}
            
            for index, subnet := range subnets {
                c.Vpn.Subnet = subnet
                c.Vpn.ClusterCIDR = ClusterCIDR[index]
                
                commons.BootstrapConsul(c.Vpn.Name, c)
                hostname := "node1"
                DHCP(c, hostname)
                hKey := fmt.Sprintf("%s/Hosts/%s/VpnAddress", c.Vpn.Name,
                    hostname)
                hkv, _, err := client.KV().Get(hKey, nil)
                commons.CheckFail(g, err)
                
                assert.NotNil(g, hkv)
                
                vKey := fmt.Sprintf("%s/TakenAddresses/%s",
                    c.Vpn.Name, string(hkv.Value))
                vkv, _, err := client.KV().Get(vKey, nil)
                commons.CheckFail(g, err)
                assert.NotNil(g, vkv)
                assert.Equal(g, string(vkv.Value), hostname)
                _, ipNet, err := net.ParseCIDR(subnet)
                commons.CheckFail(g, err)
                assert.True(g, ipNet.Contains(net.ParseIP(string(hkv.Value))))
                
                //subnet tests
                hKey = fmt.Sprintf("%s/Hosts/%s/PodSubnet", c.Vpn.Name,
                    hostname)
                hkv, _, err = client.KV().Get(hKey, nil)
                commons.CheckFail(g, err)
                
                assert.NotNil(g, hkv)
                
                vKey = fmt.Sprintf("%s/TakenPodSubnets/%s",
                    c.Vpn.Name, string(hkv.Value))
                vkv, _, err = client.KV().Get(vKey, nil)
                commons.CheckFail(g, err)
                assert.NotNil(g, vkv)
                assert.Equal(g, string(vkv.Value), hostname)
                _, ipNet, err = net.ParseCIDR(c.Vpn.ClusterCIDR)
                commons.CheckFail(g, err)
                ipPodSub, _, err := net.ParseCIDR(string(hkv.Value))
                commons.CheckFail(g, err)
                assert.True(g, ipNet.Contains(ipPodSub))
            }
        })
        
        g.It("Should preserve the ip", func() {
            g.Timeout(60 * time.Second)
            c.Vpn.Subnet = "10.65.1.0/24"
            commons.BootstrapConsul(c.Vpn.Name, c)
            d1, p1 := DHCP(c, "node1")
            d2, p2 := DHCP(c, "node1")
            assert.Equal(g, d1, d2)
            assert.Equal(g, p1, p2)
            
            d3, p3 := DHCP(c, "node1")
            d4, p4 := DHCP(c, "node1")
            assert.Equal(g, d3, d4)
            assert.Equal(g, p3, p4)
            
        })
        g.It("Should pick  all the ip from the subnet", func() {
            g.Timeout(60 * time.Second)
            
            ch := make(chan int, 2)
            c.Vpn.Subnet = "10.65.1.0/24"
            c.Vpn.ClusterCIDR = "10.10.0.0/16"
            
            commons.BootstrapConsul(c.Vpn.Name, c)
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
            
            commons.CheckFail(g, err)
            
            assert.Equal(g, len(kvs), 256)
            _, ipNet, err := net.ParseCIDR(c.Vpn.Subnet)
            commons.CheckFail(g, err)
            
            for _, kv := range kvs {
                ss := strings.Split(kv.Key, "/")
                assert.True(g, ipNet.Contains(net.ParseIP(ss[len(ss) - 1])))
                
            }
            
            // return all the allocated subnets
            kvs, _, err = client.KV().
                List(fmt.Sprintf("%s/TakenPodSubnets", c.Vpn.Name), nil)
            commons.CheckFail(g, err)
            assert.Equal(g, len(kvs), 256)
            _, ipNet, err = net.ParseCIDR(c.Vpn.ClusterCIDR)
            commons.CheckFail(g, err)
            for _, kv := range kvs {
                parts := strings.Split(kv.Key, "/")
                subn := parts[len(parts) - 2] + "/" + parts[len(parts) - 1]
                ip, _, err := net.ParseCIDR(subn)
                commons.CheckFail(g, err)
                assert.True(g, ipNet.Contains(ip))
            }
        })
        
    })
}

func TestDHCPChangeSubnet(t *testing.T) {
    g := goblin.Goblin(t)
    c := commons.Config{}
    c.Consul.Address = "localhost:8500"
    c.Consul.Scheme = "http"
    c.Vpn.Name = "tzk"
    client := commons.GetConsulClient(c)
    _, err := client.KV().DeleteTree(c.Vpn.Name, nil)
    commons.CheckFatal(err)
    
    g.Describe("DHCP", func() {
        g.It("Should not change ip if the subnet changed", func() {
            g.Timeout(30 * time.Second)
            subnet1 := "10.65.1.0/24"
            subnet2 := "10.1.0.0/16"
            _, ipNet1, err := net.ParseCIDR(subnet1)
            commons.CheckFail(g, err)
            _, ipNet2, err := net.ParseCIDR(subnet2)
            commons.CheckFail(g, err)
            c.Vpn.Subnet = subnet1
            commons.BootstrapConsul(c.Vpn.Name, c)
            newIP1, _ := DHCP(c, "node1")
            assert.True(g, ipNet1.Contains(net.ParseIP(newIP1)))
            _, err = client.KV().
                Put(&api.KVPair{Key: "tzk/Subnet", Value: []byte(subnet2)}, nil)
            commons.CheckFail(g, err)
            DHCP(c, "node1")
            newIP2, _ := DHCP(c, "node1")
            assert.Equal(g, newIP2, newIP1)
            assert.False(g, ipNet2.Contains(net.ParseIP(newIP2)))
            kv, _, err := client.KV().
                Get(fmt.Sprintf("tzk/TakenAddresses/%s", newIP1), nil)
            commons.CheckFail(g, err)
            assert.NotNil(g, kv)
        })
    })
}

func TestAssignStaticIP(t *testing.T) {
    g := goblin.Goblin(t)
    c := commons.Config{}
    c.Consul.Address = "localhost:8500"
    c.Consul.Scheme = "http"
    c.Vpn.Name = "tzk"
    c.Vpn.NodeIP = "10.187.0.50"
    client := commons.GetConsulClient(c)
    _, err := client.KV().DeleteTree(c.Vpn.Name, nil)
    commons.CheckFatal(err)
    c.Vpn.Subnet = "10.187.0.0/16"
    c.Vpn.ClusterCIDR = "10.32.0.0/12"
    c.Vpn.PodSubnet = "10.32.1.0/24"
    commons.BootstrapConsul(c.Vpn.Name, c)
    
    g.Describe("initSubnet", func() {
        g.It("Should pick the NodeIP", func() {
            hostname := "node1"
            DHCP(c, hostname)
            hKey := fmt.Sprintf("%s/Hosts/%s/VpnAddress", c.Vpn.Name, hostname)
            hkv, _, err := client.KV().Get(hKey, nil)
            commons.CheckFail(g, err)
            assert.Equal(g, string(hkv.Value), c.Vpn.NodeIP)
            
            hKey = fmt.Sprintf("%s/Hosts/%s/PodSubnet", c.Vpn.Name, hostname)
            hkv, _, err = client.KV().Get(hKey, nil)
            commons.CheckFail(g, err)
            assert.Equal(g, string(hkv.Value), c.Vpn.PodSubnet)
        })
        g.It("Should pick another ip if is taken", func() {
            hostname := "node2"
            DHCP(c, hostname)
            hKey := fmt.Sprintf("%s/Hosts/%s/VpnAddress", c.Vpn.Name, hostname)
            hkv, _, err := client.KV().Get(hKey, nil)
            commons.CheckFail(g, err)
            assert.NotEqual(g, string(hkv.Value), c.Vpn.NodeIP)
        })
        
        g.It("Should not use the ip if is not valid", func() {
            hostname := "node3"
            c.Vpn.NodeIP = "test"
            DHCP(c, hostname)
            hKey := fmt.Sprintf("%s/Hosts/%s/VpnAddress", c.Vpn.Name, hostname)
            hkv, _, err := client.KV().Get(hKey, nil)
            commons.CheckFail(g, err)
            assert.NotEqual(g, string(hkv.Value), c.Vpn.NodeIP)
        })
        
        g.It("Should not use the ip if is in the current subnet", func() {
            hostname := "node5"
            c.Vpn.NodeIP = "10.32.3.3"
            DHCP(c, hostname)
            hKey := fmt.Sprintf("%s/Hosts/%s/VpnAddress", c.Vpn.Name, hostname)
            hkv, _, err := client.KV().Get(hKey, nil)
            commons.CheckFail(g, err)
            assert.NotEqual(g, string(hkv.Value), c.Vpn.NodeIP)
        })
    })
}

func TestInitSubnet(t *testing.T) {
    g := goblin.Goblin(t)
    c := commons.Config{}
    c.Consul.Address = "localhost:8500"
    c.Consul.Scheme = "http"
    c.Vpn.Name = "tzk"
    client := commons.GetConsulClient(c)
    
    g.Describe("initSubnet", func() {
        g.It("Should assign the default subnet", func() {
            _, err := client.KV().DeleteTree(c.Vpn.Name, nil)
            if err != nil {
                log.Fatal(err)
            }
            InitSubnet(c)
            kv, _, err := client.KV().Get(fmt.Sprintf("%s/Subnet", c.Vpn.Name),
                nil)
            commons.CheckFail(g, err)
            assert.NotNil(g, kv)
            assert.Equal(g, string(kv.Value), DefaultSubnet)
            
            kv, _, err = client.KV().
                Get(fmt.Sprintf("%s/ClusterCIDR", c.Vpn.Name),
                nil)
            commons.CheckFail(g, err)
            assert.NotNil(g, kv)
            assert.Equal(g, string(kv.Value), DefaultClusterCIDR)
        })
        
        g.It("Should use the user defined subnet and pod cidr", func() {
            _, err := client.KV().DeleteTree(c.Vpn.Name, nil)
            if err != nil {
                log.Fatal(err)
            }
            c.Vpn.ClusterCIDR = "10.81.0.0/16"
            c.Vpn.Subnet = "172.2.0.0/16"
            
            InitSubnet(c)
            kv, _, err := client.KV().Get(fmt.Sprintf("%s/Subnet", c.Vpn.Name),
                nil)
            commons.CheckFail(g, err)
            assert.NotNil(g, kv)
            assert.Equal(g, string(kv.Value), "172.2.0.0/16")
            
            kv, _, err = client.KV().
                Get(fmt.Sprintf("%s/ClusterCIDR", c.Vpn.Name), nil)
            commons.CheckFail(g, err)
            assert.NotNil(g, kv)
            assert.Equal(g, string(kv.Value), "10.81.0.0/16")
        })
        
        g.It("Should not change subnet if already exist", func() {
            _, err := client.KV().DeleteTree(c.Vpn.Name, nil)
            if err != nil {
                log.Fatal(err)
            }
            kv := &api.KVPair{Key: fmt.Sprintf("%s/Subnet", c.Vpn.Name),
                Value: []byte("10.1.0.0/16")}
            _, err = client.KV().Put(kv, nil)
            commons.CheckFail(g, err)
            InitSubnet(c)
            kv, _, err = client.KV().Get(fmt.Sprintf("%s/Subnet", c.Vpn.Name),
                nil)
            commons.CheckFail(g, err)
            assert.NotEqual(g, string(kv.Value), DefaultSubnet)
        })
    })
}

func TestGetSubnets(t *testing.T) {
    g := goblin.Goblin(t)
    c := commons.Config{}
    c.Consul.Address = "localhost:8500"
    c.Consul.Scheme = "http"
    c.Vpn.Name = "tzk"
    
    g.Describe("GetSubnets", func() {
        g.It("Should return all the subnets with /24 netmask from the cidr",
            func() {
                subnets := getSubnets("10.32.0.0/12", []string{})
                assert.Equal(g, len(subnets), 4096)
                for _, i := range subnets {
                    _, _, err := net.ParseCIDR(i)
                    assert.Nil(g, err)
                }
            })
        g.It("Should not return taken ones",
            func() {
                subnets := getSubnets("10.32.0.0/12",
                    []string{"10.32.1.0/24", "10.33.1.0/24"})
                assert.Equal(g, len(subnets), 4096 - 2)
                for _, i := range subnets {
                    _, _, err := net.ParseCIDR(i)
                    assert.Nil(g, err)
                }
            })
    })
}

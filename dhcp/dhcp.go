package dhcp

import (
    "fmt"
    "math/rand"
    "net"
    "tzk-daemon/commons"
    "time"
    
    log "github.com/Sirupsen/logrus"
    "github.com/hashicorp/consul/api"
    "strings"
)

//DefaultSubnet the default cidr of the vpn
const DefaultSubnet = "192.168.0.0/16"

//DefaultClusterCIDR the cidr of the pods
const DefaultClusterCIDR = "10.32.0.0/12"

func inc(ip net.IP) {
    for j := len(ip) - 1; j >= 0; j-- {
        ip[j]++
        if ip[j] > 0 {
            break
        }
    }
}

func getSubnets(cidr string, takenSubnets[]string) []string {
    subnets := []string{}
    ip, ipNet, err := net.ParseCIDR(cidr)
    
    if err != nil {
        log.Fatal(err)
    }
    // get first
    j := len(ip) - 2
    ns := ip.String() + "/24"
    if !commons.StringSliceContains(takenSubnets, ns) {
        subnets = append(subnets, ns)
    }
    for {
        ip[j]++
        
        if ip[j] == 0 {
            ip[j - 1]++
        }
        
        if ipNet.Contains(ip) {
            ns = ip.String() + "/24"
            if !commons.StringSliceContains(takenSubnets, ns) {
                subnets = append(subnets, ns)
            }
        } else {
            break
        }
    }
    return subnets
}

func cidrAddress(cidr string, TakenAddresses []string) []string {
    ips := []string{}
    
    ip, ipNet, err := net.ParseCIDR(cidr)
    
    if err != nil {
        log.Fatal(err)
    }
    
    for ip = ip.Mask(ipNet.Mask); ipNet.Contains(ip); inc(ip) {
        if !commons.StringSliceContains(TakenAddresses, ip.String()) {
            ips = append(ips, ip.String())
        }
    }
    return ips
}

func currentIP(client *api.Client, c commons.Config, hostname string) *string {
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

func currentCIDR(client *api.Client, c commons.Config,
hostname string) *string {
    var ret string
    // check if hostname has a ip associate with it
    kv, _, err := client.KV().
        Get(fmt.Sprintf("%s/Hosts/%s/PodSubnet", c.Vpn.Name, hostname), nil)
    if err != nil {
        log.Fatal(err)
    }
    if kv != nil {
        ret = string(kv.Value)
        return &ret
    }
    return nil
}

func assignCIDR(client *api.Client, c commons.Config, hostname string) string {
    ci := currentCIDR(client, c, hostname)
    if ci != nil {
        //check ip is register in TakenAddresses prefix
        kv, _, err := client.KV().
            Get(fmt.Sprintf("%s/TakenPodSubnets/%s", c.Vpn.Name, *ci), nil)
        commons.CheckFatal(err)
        if kv != nil {
            if string(kv.Value) == hostname {
                return *ci
            }
        }
    }
    // return pod Cidr
    kv, _, err := client.KV().
        Get(fmt.Sprintf("%s/ClusterCIDR", c.Vpn.Name), nil)
    commons.CheckFatal(err)
    ClusterCIDR := string(kv.Value)
    
    // return all the allocated subnets
    kvs, _, err := client.KV().
        List(fmt.Sprintf("%s/TakenPodSubnets", c.Vpn.Name), nil)
    
    if err != nil {
        log.Fatal(err)
    }
    return takeSubnet(kvs, ClusterCIDR, client, c, hostname)
}

func assignIP(client *api.Client, c commons.Config, hostname string) string {
    
    ci := currentIP(client, c, hostname)
    if ci != nil {
        //check ip is register in TakenAddresses prefix
        kv, _, err := client.KV().
            Get(fmt.Sprintf("%s/TakenAddresses/%s", c.Vpn.Name, *ci), nil)
        commons.CheckFatal(err)
        if kv != nil {
            if string(kv.Value) == hostname {
                return *ci
            }
        }
    }
    // return subnet
    kv, _, err := client.KV().Get(fmt.Sprintf("%s/Subnet", c.Vpn.Name), nil)
    commons.CheckFatal(err)
    subnet := string(kv.Value)
    
    // return all the allocated addresses
    kvs, _, err := client.KV().
        List(fmt.Sprintf("%s/TakenAddresses", c.Vpn.Name), nil)
    
    if err != nil {
        log.Fatal(err)
    }
    return takeAddress(kvs, subnet, client, c, hostname)
}

func takeAddress(kvs api.KVPairs, subnet string, client *api.Client,
c commons.Config, hostname string) string {
    rand.Seed(time.Now().UnixNano())
    var kv *api.KVPair
    newIP := c.Vpn.NodeIP
    // check if the ip is ok
    ip := net.ParseIP(newIP)
    if ip != nil {
        _, ipn, err := net.ParseCIDR(subnet)
        commons.CheckFatal(err)
        if ipn.Contains(ip) {
            kv = &api.KVPair{}
            kv.CreateIndex = 0
            kv.Key = fmt.Sprintf("%s/TakenAddresses/%s", c.Vpn.Name, newIP)
            kv.Value = []byte(hostname)
            //save new ip and lock it
            ok, _, err := client.KV().CAS(kv, nil)
            commons.CheckFatal(err)
            
            if ok {
                return newIP
            }
        }
    }
    
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
        
        kv = &api.KVPair{}
        kv.CreateIndex = 0
        kv.Key = fmt.Sprintf("%s/TakenAddresses/%s", c.Vpn.Name, newIP)
        kv.Value = []byte(hostname)
        
        //save new ip and lock it
        ok, _, err := client.KV().CAS(kv, nil)
        commons.CheckFatal(err)
        
        if ok {
            return newIP
        }
        takenAddresses = append(takenAddresses, newIP)
        time.Sleep(100 * time.Millisecond)
        log.Info("Requesting new IP")
    }
}

func takeSubnet(kvs api.KVPairs, ClusterCIDR string, client *api.Client,
c commons.Config, hostname string) string {
    rand.Seed(time.Now().UnixNano())
    var kv *api.KVPair
    newSubnet := c.Vpn.PodSubnet
    if len(newSubnet) > 0 {
        // check if the subnet is ok
        ip, ipNet, err := net.ParseCIDR(newSubnet)
        if err != nil {
            log.Error("You picked a bad pod subnet for this host")
            commons.CheckFatal(err)
        }
        mask := strings.Split(ipNet.String(), "/")
        if mask[1] != "24" {
            log.Fatal("Net mask should be 24 ->", mask[1])
            commons.CheckFatal(err)
        }
        
        _, ipn, err := net.ParseCIDR(ClusterCIDR)
        commons.CheckFatal(err)
        if ! ipn.Contains(ip) {
            log.Fatal("The subnet pod that you picked is not" +
                " in the ClusterCIDR range: ", newSubnet, ",", ClusterCIDR)
        }
        newSubnet = ipNet.String()
        
        kv = &api.KVPair{}
        kv.CreateIndex = 0
        kv.Key = fmt.Sprintf("%s/TakenPodSubnets/%s", c.Vpn.Name, newSubnet)
        kv.Value = []byte(hostname)
        //save new ip and lock it
        ok, _, err := client.KV().CAS(kv, nil)
        commons.CheckFatal(err)
        if ok {
            return newSubnet
        }
        
    }
    
    takenSubnets := []string{}
    for _, kv := range kvs {
        takenSubnets = append(takenSubnets, kv.Key)
    }
    
    for {
        availableSubnets := getSubnets(ClusterCIDR, takenSubnets)
        if len(availableSubnets) == 0 {
            log.Fatal("Not Subnets to take")
        }
        newSubnet = availableSubnets[rand.Intn(len(availableSubnets))]
        
        kv = &api.KVPair{}
        kv.CreateIndex = 0
        kv.Key = fmt.Sprintf("%s/TakenPodSubnets/%s", c.Vpn.Name, newSubnet)
        kv.Value = []byte(hostname)
        
        //save new subnet and lock it
        ok, _, err := client.KV().CAS(kv, nil)
        commons.CheckFatal(err)
        
        if ok {
            return newSubnet
        }
        takenSubnets = append(takenSubnets, newSubnet)
        time.Sleep(100 * time.Millisecond)
        log.Info("Requesting new newSubnet")
    }
}


//DHCP pick a ip address from the subnet cidr
//use the locking features of consul to guarantee
//ip per node
func DHCP(c commons.Config, hostname string) (addr string, podSubnet string) {
    client := commons.GetConsulClient(c)
    address := assignIP(client, c, hostname)
    subnet := assignCIDR(client, c, hostname)
    //store ip address in the host prefix
    kv := &api.KVPair{}
    kv.Key = fmt.Sprintf("%s/Hosts/%s/VpnAddress", c.Vpn.Name, hostname)
    kv.Value = []byte(address)
    _, err := client.KV().Put(kv, nil)
    commons.CheckFatal(err)
    
    kv.Key = fmt.Sprintf("%s/Hosts/%s/PodSubnet", c.Vpn.Name, hostname)
    kv.Value = []byte(subnet)
    _, err = client.KV().Put(kv, nil)
    commons.CheckFatal(err)
    
    log.Infof("The ip address %s have been assigned to this node", address)
    log.Infof("The pod subnet  %s have been assigned to this node", subnet)
    return address, subnet
}
//InitSubnet ...
func InitSubnet(c commons.Config) {
    client := commons.GetConsulClient(c)
    kv, _, err := client.KV().Get(fmt.Sprintf("%s/Subnet", c.Vpn.Name), nil)
    commons.CheckFatal(err)
    s := DefaultSubnet
    if c.Vpn.Subnet != "" {
        s = c.Vpn.Subnet
    }
    if kv == nil {
        kv = &api.KVPair{Key: fmt.Sprintf("%s/Subnet", c.Vpn.Name),
            Value: []byte(s)}
        _, err = client.KV().Put(kv, nil)
        commons.CheckFatal(err)
        
    }
    
    kv, _, err = client.KV().Get(fmt.Sprintf("%s/ClusterCIDR", c.Vpn.Name), nil)
    commons.CheckFatal(err)
    ps := DefaultClusterCIDR
    if c.Vpn.ClusterCIDR != "" {
        ps = c.Vpn.ClusterCIDR
    }
    if kv == nil {
        kv = &api.KVPair{Key: fmt.Sprintf("%s/ClusterCIDR", c.Vpn.Name),
            Value: []byte(ps)}
        _, err = client.KV().Put(kv, nil)
        commons.CheckFatal(err)
        
    }
    
}

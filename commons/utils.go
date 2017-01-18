package commons

import (
    log "github.com/Sirupsen/logrus"
    "github.com/hashicorp/consul/api"
    "runtime"
    "io/ioutil"
    "github.com/criloz/goblin"
    "fmt"
)

//GetConsulClient Return a new consul client
func GetConsulClient(c Config) *api.Client {
    ac := &api.Config{Address: c.Consul.Address,
        Scheme: c.Consul.Scheme,
        Token:  c.Consul.Token}
    client, err := api.NewClient(ac)
    CheckFatal(err)
    return client
}

//CheckFatal close the app and show where the error comes from
func CheckFatal(e error) {
    _, file, line, _ := runtime.Caller(1)
    if e != nil {
        log.WithFields(log.Fields{"file": file,
            "line": line,
        }).Panic(e)
    }
}

//StringSliceContains check if a string slice contains the searchString
func StringSliceContains(stringSlice []string, searchString string) bool {
    for _, value := range stringSlice {
        if value == searchString {
            return true
        }
    }
    return false
}

//BootstrapConsul useful function just for testing
func BootstrapConsul(vpnName string, c Config) {
    client := GetConsulClient(c)
    if len(c.Vpn.ClusterCIDR)== 0{
        c.Vpn.ClusterCIDR = "10.32.0.0/12"
    }
    _, err := client.KV().DeleteTree(vpnName, nil)
    CheckFatal(err)
    kp := &api.KVPair{Key: fmt.Sprintf("%s/Subnet", vpnName),
        Value: []byte(c.Vpn.Subnet)}
    
    _, err = client.KV().Put(kp, nil)
    CheckFatal(err)
    kp = &api.KVPair{Key: fmt.Sprintf("%s/ClusterCIDR", vpnName),
        Value: []byte(c.Vpn.ClusterCIDR)}
    
    _, err = client.KV().Put(kp, nil)
    CheckFatal(err)
    
    c.Vpn.PublicKeyFile = "/tmp/pubkey"
    err = ioutil.WriteFile(c.Vpn.PublicKeyFile, []byte("pubkey = pubkey"), 0755)
    
    CheckFatal(err)
    
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
//CheckFail   for test
func CheckFail(g *goblin.G, err error) {
    if err != nil {
        g.Fail(err)
    }
}

package main

import (
    "testing"
    "tzk-daemon/commons"
)

func TestGetPing(t *testing.T){
    c := commons.Config{}
    c.Consul.Address = "localhost:8500"
    c.Consul.Scheme = "http"
    c.Vpn.Name = "tzk"
    c.Vpn.PublicKeyFile = "/tmp/pk"
    c.Vpn.Subnet = "10.187.0.0/16"
    c.Vpn.NodeIP = ""
    c.Vpn.ClusterCIDR = "10.32.0.0/12"
    c.Vpn.PodSubnet = ""
    c.Consul.Address = "kv.vpn12.nebtex.com"
    c.Consul.Scheme = "https"
    c.Consul.Token = "1c497c6a-d703-487c-8b74-edd94d2fb0a9"
    
    getIP(c)
}


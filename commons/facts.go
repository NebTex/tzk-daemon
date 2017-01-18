package commons

import (
    log "github.com/Sirupsen/logrus"
    "net"
    "os/exec"
    "os"
)

//Facts store useful info about the node
type Facts struct {
    Addresses     map[string]string
    HasChanged    bool `json:"-"`
    Container     string
    City          string
    CountryCode   string
    CountryName   string
    RegionCode    string
    RegionName    string
    ZipCode       string
    TimeZone      string
    MetroCode     int
    Latitude      float32
    Longitude     float32
    ContinentCode string
    PublicKey     string
    Hostname      string
}

//Dumps contain the results of tinc dump commands
type Dumps struct {
    Nodes       string
    Edges       string
    Subnets     string
    Connections string
    Graph       string
    Invitations string
}

//Get the dump commands output
func (d *Dumps) Get(c Config) {
    if d == nil {
        d = &Dumps{}
    }
    out, err := exec.Command("/usr/sbin/tinc", "-n", c.Vpn.Name,
        "dump", "nodes").Output()
    if err != nil {
        log.Error(err)
    }
    d.Nodes = string(out)
    
    out, err = exec.Command("/usr/sbin/tinc", "-n", c.Vpn.Name,
        "dump", "edges").Output()
    if err != nil {
        log.Error(err)
    }
    d.Edges = string(out)
    
    out, err = exec.Command("/usr/sbin/tinc", "-n", c.Vpn.Name,
        "dump", "subnets").
        Output()
    if err != nil {
        log.Error(err)
    }
    d.Subnets = string(out)
    
    out, err = exec.Command("/usr/sbin/tinc", "-n", c.Vpn.Name,
        "dump", "connections").
        Output()
    if err != nil {
        log.Error(err)
    }
    d.Connections = string(out)
    
    out, err = exec.Command("/usr/sbin/tinc", "-n", c.Vpn.Name,
        "dump", "graph").Output()
    if err != nil {
        log.Error(err)
    }
    d.Graph = string(out)
    
    out, err = exec.Command("/usr/sbin/tinc", "-n", c.Vpn.Name,
        "dump", "invitations").
        Output()
    if err != nil {
        log.Error(err)
    }
    d.Invitations = string(out)
}


//AddAddress add ipv4 or ipv6 address to the Fact map
//can only add global unicast address
func (f *Facts) AddAddress(addr string) {
    
    ip := net.ParseIP(addr)
    if ip == nil {
        log.Errorf("%s is not a valid Address\n", addr)
    }
    if ip.IsGlobalUnicast() {
        if f.Addresses == nil {
            f.Addresses = make(map[string]string)
        }
        if _, ok := f.Addresses[addr]; !ok {
            f.Addresses[addr] = ""
            f.HasChanged = true
        }
        
    }
}

//GetContainerStatus ...
func (f *Facts) GetContainerStatus() {
    if f.Container != os.Getenv("container") {
        f.HasChanged = true
    }
    f.Container = os.Getenv("container")
}






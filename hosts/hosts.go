package hosts

import (
    "fmt"
    log "github.com/Sirupsen/logrus"
    "io/ioutil"
    "regexp"
    "strings"
    "tzk-daemon/commons"
)

//Hosts list of Host
type Hosts map[string]commons.Host
//Vpn all the vpn info
type Vpn struct {
    Hosts  Hosts
    Subnet string
    ClusterCIDR string
}



//SetHostFile maintain updates the host file
func (v *Vpn) SetHostFile(thisHostName string) {
    //read file
    d, err := ioutil.ReadFile("/etc/hosts")
    commons.CheckFatal(err)
    //replace
    th, ok := v.Hosts[thisHostName]
    if !ok {
        log.Fatal("This host is not defined on consul")
        return
    }
    o := v.Hosts.manageHostBlock(string(d), th)
    //save
    err = ioutil.WriteFile("/etc/hosts", []byte(o), 0644)
    commons.CheckFatal(err)
}

func (hs Hosts) manageHostBlock(input string, thisHost commons.Host) string {
    re := regexp.MustCompile(`(?ms)#\/tzk\/NoEdit(.)*#\/tzk\/NoEdit`)
    all := re.FindAllString(input, -1)
    if len(all) == 0 {
        return fmt.Sprintf(`%s
#/tzk/NoEdit
%s
#/tzk/NoEdit`, input, hs.parseToFileFormat(thisHost))
        
    }
    return re.ReplaceAllString(input, fmt.Sprintf(`#/tzk/NoEdit
%s
#/tzk/NoEdit`, hs.parseToFileFormat(thisHost)))
}

func (hs Hosts) parseToFileFormat(thisHost commons.Host) string {
    hFile := make([]string, len(hs))
    for _, h := range hs {
        entry := fmt.Sprintf("%s\t%s.%s.local",
            h.VpnAddress, h.Facts.Hostname, "tzk")
        hFile = append(hFile, entry)
    }
    // make possible to run weave network with kubernetes
    entry := fmt.Sprintf("%s\t%s", thisHost.VpnAddress, thisHost.Facts.Hostname)
    hFile = append(hFile, entry)
    
    return strings.Join(hFile, "\n")
}

//FixName ...
func FixName(s string) string {
    re := regexp.MustCompile(`[^A-Za-z0-9_]`)
    return re.ReplaceAllString(s, "_")
}

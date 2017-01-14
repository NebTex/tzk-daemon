package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"os/exec"

	log "github.com/Sirupsen/logrus"
)

//GetTincInfo get hostname and public key of the node
func (f *Facts) GetTincInfo(c Config, Hostname func() (string, error)) {
	// get public key
	publicKeyPath := c.Vpn.PublicKeyFile
	data, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		log.Fatal(err)
		return
	}
	ss := strings.Split(string(data), "=")
	if len(ss) < 2 {
		log.Fatal("tinc public key has bad format")
	}
	pKey := strings.TrimSpace(ss[1])
	if f.PublicKey != pKey {
		f.HasChanged = true
	}
	f.PublicKey = pKey
	hn, err := Hostname()
	if err != nil {
		log.Fatal("Cant read the hostname of this node")
		log.Fatal(err)
		return
	}
	if f.Hostname != hn {
		f.HasChanged = true
	}
	f.Hostname = hn
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

//Host contain all the node info
type Host struct {
	VpnAddress string
	Facts      Facts
	Configs    map[string]string
	Dumps      *Dumps
}

//Hosts list of Host
type Hosts map[string]Host

//Vpn all the vpn info
type Vpn struct {
	Hosts  Hosts
	Subnet string
}

//Files contains all the files needed to configure tinc and
//connect to other nodes
type Files struct {
	Hosts map[string]string
	Tinc  map[string]string
}

//GenerateFiles create the tinc conf files,
// and host files from the Vpn struct
func (v *Vpn) GenerateFiles(thisHostname string, c Config) *Files {
	log.Infof("Generating files for %s", thisHostname)
	f := &Files{}
	f.Hosts = make(map[string]string)
	f.Tinc = make(map[string]string)

	th, ok := v.Hosts[thisHostname]
	if !ok {
		log.Fatal("This host is not defined on consul")
		return nil
	}

	// create tinc up
	f.Tinc["tinc-up"] = fmt.Sprintf(`#!/bin/sh
ip link set $INTERFACE up
ip addr add  %s dev $INTERFACE
ip route add %s dev $INTERFACE`, th.VpnAddress, v.Subnet)

	// create tinc down
	f.Tinc["tinc-down"] = fmt.Sprintf(`#!/bin/sh
ip route del %s dev $INTERFACE
ip addr del %s dev $INTERFACE
ip link set $INTERFACE down`, v.Subnet, th.VpnAddress)

	cFile := []string{}
	keys := make([]string, 0, len(th.Configs)+1)
	for k := range th.Configs {
		keys = append(keys, k)
	}
	keys = append(keys, "Name")
	sort.Strings(keys)
	// create tinc conf
	for _, k := range keys {
		if k == "Name" {
			cFile = append(cFile, fmt.Sprintf("%s=%s", k, fixName(thisHostname)))
			continue
		}
		cFile = append(cFile, fmt.Sprintf("%s=%s", k, th.Configs[k]))
	}

	f.Tinc["tinc.conf"] = strings.Join(cFile, "\n")
	// create host files
	for _, host := range v.Hosts {

		hostFile := []string{}
		hostFile = append(hostFile,
			fmt.Sprintf("Ed25519PublicKey=%s", host.Facts.PublicKey))
		keys := make([]string, 0, len(host.Facts.Addresses))
		for k := range host.Facts.Addresses {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, address := range keys {
			hostFile = append(hostFile,
				fmt.Sprintf("Address=%s", address))
		}
		hostFile = append(hostFile, fmt.Sprintf("Subnet=%s", host.VpnAddress))
		f.Hosts[fixName(host.Facts.Hostname)] = strings.Join(hostFile, "\n")
	}

	return f
}

func mapToString(m map[string]string, strL []string) []string {
	ret := strL
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		ret = append(ret, k)
		ret = append(ret, m[k])

	}
	return ret

}

//ToString write all the files to a string, it should return always the same
// string with the same input
func (f *Files) ToString() string {
	allFiles := []string{}
	//Tinc file to string
	allFiles = mapToString(f.Tinc, allFiles)
	//Host files to sting
	allFiles = mapToString(f.Hosts, allFiles)
	return strings.Join(allFiles, "\n")
}

//Equal check if two files have the same info
func (f *Files) Equal(f2 *Files) bool {
	if f2 == nil {
		return f == nil
	}
	return f.ToString() == f2.ToString()
}

func saveMap(m map[string]string, folder string) error {
	var permission os.FileMode
	for k, v := range m {
		switch k {
		case "tinc.conf":
			permission = 0640
		case "tinc-up":
			permission = 0755
		case "tinc-down":
			permission = 0755
		default:
			permission = 0644
		}
		err := ioutil.WriteFile(filepath.Join(folder, k), []byte(v), permission)
		if err != nil {
			return err
		}
	}
	return nil
}

//Write save the generated files
func (f *Files) Write(c Config) {
	err := saveMap(f.Tinc, fmt.Sprintf("/etc/tinc/%s/", c.Vpn.Name))
	if err != nil {
		log.Error("Could not save the tinc config files")
		log.Fatal(err)
	}
	hosts := fmt.Sprintf("/etc/tinc/%s/hosts/", c.Vpn.Name)
	err = os.RemoveAll(hosts)
	if err != nil {
		log.Error("Failed to remove hosts files")
		log.Fatal(err)
	}
	err = os.MkdirAll(hosts, 644)
	if err != nil {
		log.Error("Failed to create hosts path")
		log.Fatal(err)
	}
	err = saveMap(f.Hosts, hosts)
	if err != nil {
		log.Error("Could not save the hosts  files")
		log.Fatal(err)
	}
}

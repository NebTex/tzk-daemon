package main

import (
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
)

//GetTincInfo get hostname and public key of the node
func (f *Facts) GetTincInfo(c Config) {
	// get public key
	publicKeyPath := filepath.Join("/etc/tinc/", c.Vpn.Name, "key.pub")
	data, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		log.Fatal(err)
		return
	}
	pKey := string(data)
	if f.PublicKey != pKey {
		f.HasChanged = true
	}
	f.PublicKey = pKey
	hn, err := os.Hostname()
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

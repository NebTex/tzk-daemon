package commons

import (
	"io/ioutil"
	"strings"
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


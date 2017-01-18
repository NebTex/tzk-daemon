package hosts

import (

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/consul/api"
	"github.com/mitchellh/consulstructure"
    "tzk-daemon/commons"
)
/*
func toByte(value interface{}) []byte {
	switch v := value.(type) {
	case string:
		return []byte(v)
	case int:
		return []byte(strconv.Itoa(v))
	case float32:
		return []byte(strconv.FormatFloat(float64(v), 'f', 4, 32))
	default:
	}
	return []byte("")
}*/


//WatchConsul watch  consul for changes in the vpn info
//call f each time that a change is detected
func WatchConsul(c commons.Config, f func(v *Vpn, close func())) {
	// Create our decoder
	closeCh := make(chan bool, 1)
	updateCh := make(chan interface{})
	errCh := make(chan error)
	ac := &api.Config{Address: c.Consul.Address,
		Scheme: c.Consul.Scheme,
		Token:  c.Consul.Token}
	decoder := &consulstructure.Decoder{
		Target:   &Vpn{},
		Consul:   ac,
		Prefix:   c.Vpn.Name,
		UpdateCh: updateCh,
		ErrCh:    errCh,
	}

	// Run the decoder and wait for changes
	go decoder.Run()

	closeT := func() {
		closeCh <- true
		err := decoder.Close()
		commons.CheckFatal(err)
	}

	for {
		select {
		case v := <-updateCh:
			f(v.(*Vpn), closeT)
		case err := <-errCh:
			log.Error(err)
		case <-closeCh:
			return
		}
	}

}

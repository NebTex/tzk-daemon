package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/consul/api"
	"github.com/mitchellh/consulstructure"
	"strconv"
)

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
}

func getConsulClient(c Config) *api.Client {
	ac := &api.Config{Address: c.Consul.Address,
		Scheme: c.Consul.Scheme,
		Token:  c.Consul.Token}
	client, err := api.NewClient(ac)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func (f *Facts) addKey(k string, v interface{}) *api.KVTxnOp {
	return &api.KVTxnOp{Verb: "set",
		Key:   fmt.Sprintf("tzk/Hosts/%s/Facts/%s", f.Hostname, k),
		Value: toByte(v)}
}

func (f *Facts) parseAddress(txn api.KVTxnOps) api.KVTxnOps {
	for address := range f.Addresses {
		txn = append(txn, f.addKey(fmt.Sprintf("Addresses/%s", address), ""))
	}
	return txn
}

//SendToConsul save facts in consul
func (f *Facts) SendToConsul(c Config) {
	if !f.HasChanged {
		return
	}

	ctx := api.KVTxnOps{f.addKey("Container", f.Container),
		f.addKey("City", f.City),
		f.addKey("CountryCode", f.CountryCode),
		f.addKey("RegionCode", f.RegionCode),
		f.addKey("RegionName", f.RegionName),
		f.addKey("ZipCode", f.ZipCode),
		f.addKey("TimeZone", f.TimeZone),
		f.addKey("MetroCode", f.MetroCode),
		f.addKey("Latitude", f.Latitude),
		f.addKey("Longitude", f.Longitude),
		f.addKey("ContinentCode", f.ContinentCode),
		f.addKey("PublicKey", f.PublicKey),
		f.addKey("HostName", f.Hostname)}
	ctx = f.parseAddress(ctx)
	log.Infoln("Storing info on Consul ...")
	//set new values
	client := getConsulClient(c)
	ok, _, _, err := client.KV().Txn(ctx, nil)
	if err != nil || !ok {
		log.Errorln("Failed to make a request to the consul service")
		log.Fatal(err)
		return
	}
	f.HasChanged = false
}

//WatchConsul watch  consul for changes in the vpn info
//call f each time that a change is detected
func WatchConsul(c Config, f func(v *Vpn, close func())) {
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
		checkFatal(err)
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

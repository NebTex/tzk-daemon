package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/consul/api"
)

func (h *Host) addConfig(tx api.KVTxnOps, key string,
	value string) api.KVTxnOps {
	k := fmt.Sprintf("tzk/Hosts/%s/Configs/%s", h.Facts.Hostname, key)
	kvtx := &api.KVTxnOp{Key: k, Value: []byte(value), Verb: "set"}
	return append(tx, kvtx)
}

//Config store the config info
type Config struct {
	Vpn struct {
		Name          string
		PublicKeyFile string
	}
	Consul struct {
		Address string
		Token   string
		Scheme  string
	}
}

//SetConfigConsul save tinc configuration of the node on consul
func (h *Host) SetConfigConsul(c Config) {
	client := getConsulClient(c)
	//check if config already exist
	kvs, _, err := client.KV().
		List(fmt.Sprintf("%s/Hosts/%s/Configs", c.Vpn.Name, h.Facts.Hostname), nil)
	checkFatal(err)
	//if configs exist do nothing
	if len(kvs) > 0 {
		log.Info("using  existing configs")
		return
	}
	//set default config otherwise
	log.Info("set default configs")
	configs := api.KVTxnOps{}
	configs = h.addConfig(configs, "AutoConnect", "yes")
	configs = h.addConfig(configs, "AddressFamily", "ipv4")
	configs = h.addConfig(configs, "Device", "/dev/net/tun")
	configs = h.addConfig(configs, "LocalDiscovery", "yes")
	configs = h.addConfig(configs, "MaxTimeout", "300")
	configs = h.addConfig(configs, "Mode", "router")
	_, _, _, err = client.KV().Txn(configs, nil)
	checkFatal(err)
}

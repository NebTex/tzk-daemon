package commons

import (
	"github.com/criloz/goblin"
	"github.com/stretchr/testify/assert"
	"testing"
)

//TestDHCP
func TestHost_SetConfigConsul(t *testing.T) {
	g := goblin.Goblin(t)
	c := Config{}
	c.Consul.Address = "localhost:8500"
	c.Consul.Scheme = "http"
	c.Vpn.Name = "tzk"
	client := GetConsulClient(c)

	g.Describe("DHCP", func() {
		g.It("Should set a new config", func() {
            c.Vpn.Subnet = "10.0.0.0/16"
			BootstrapConsul(c.Vpn.Name, c)
			h := Host{}
			h.Facts.Hostname = "node1"
			h.SetConfigConsul(c)
			kvs, _, err := client.KV().List("tzk/Hosts/node1/Configs", nil)
			CheckFail(g, err)
			assert.Equal(g, len(kvs), 6)
		})

		g.It("Should not change the configs if they exists", func() {
            c.Vpn.Subnet = "10.0.0.0/16"
			BootstrapConsul(c.Vpn.Name, c)
			h := Host{}
			h.Facts.Hostname = "node1"
			h.SetConfigConsul(c)
			_, err := client.KV().Delete("tzk/Hosts/node1/Configs/Mode", nil)
			CheckFatal(err)
			h.SetConfigConsul(c)
			kvs, _, err := client.KV().List("tzk/Hosts/node1/Configs", nil)
			CheckFail(g, err)
			assert.Equal(g, len(kvs), 5)
		})

	})

}



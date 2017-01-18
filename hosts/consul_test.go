package hosts

import (
	"fmt"
	"github.com/criloz/goblin"
	"github.com/hashicorp/consul/api"
	"testing"
	"time"
    "tzk-daemon/commons"
    "tzk-daemon/dhcp"
)

//TestDHCP
func TestWatchConsul(t *testing.T) {
	g := goblin.Goblin(t)
	c := commons.Config{}
	c.Consul.Address = "localhost:8500"
	c.Consul.Scheme = "http"
	c.Vpn.Name = "tzk"
	client := commons.GetConsulClient(c)

	g.Describe("WatchConsul", func() {
		g.It("Check if the changes are feed to each daemon",
			func(done goblin.Done) {
				g.Timeout(10 * time.Second)
				count := 0
				handle := func(v *Vpn, close func()) {
					fmt.Print(v)
					count++
					fmt.Println("count: ", count)

					if count >= 3 {
						close()
						done()
					}
				}
                c.Vpn.Subnet = "10.1.0.0/16"
				commons.BootstrapConsul(c.Vpn.Name, c)
				dhcp.DHCP(c, "node1")
				go WatchConsul(c, handle)
				time.Sleep(1 * time.Second)
				_, err := client.KV().
					Put(&api.KVPair{Key: "tzk/Subnet",
						Value: []byte("10.100.0.0/16")}, nil)
				if err != nil {
					g.Fail(err)
				}
				time.Sleep(1 * time.Second)
				_, err = client.KV().
					Put(&api.KVPair{Key: "tzk/Subnet",
						Value: []byte("10.70.0.0/16")}, nil)
				if err != nil {
					g.Fail(err)
				}

			})

	})
}

package main

import (
	"github.com/criloz/goblin"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

//TestParseToFileFormat
func TestParseToFileFormat(t *testing.T) {
	g := goblin.Goblin(t)
	c := Config{}
	c.Consul.Address = "localhost:8500"
	c.Consul.Scheme = "http"
	c.Vpn.Name = "tzk"
	g.Describe("ParseToFileFormat", func() {
		g.It("should insert all the nodes in the host file",
			func(done goblin.Done) {
				handle := func(v *Vpn, close func()) {
					o := v.Hosts.parseToFileFormat()
					assert.True(g, strings.Contains(o, "node3.tzk.local"))
					assert.True(g, strings.Contains(o, "node4.tzk.local"))
					assert.True(g, strings.Contains(o, "node1.tzk.local"))
					assert.True(g, strings.Contains(o, "node2.tzk.local"))
					close()
					done()
				}
				subnet1 := "10.65.1.0/24"
				bootstrapConsul(subnet1, c.Vpn.Name, c)
				DHCP(c, "node1")
				addHost(c, "node2", "pubkey2", "185.36.25.14")
				addHost(c, "node3", "pubkey3", "108.36.25.14")
				addHost(c, "node4", "pubkey4", "95.36.25.14")
				go WatchConsul(c, handle)

			})

	})
}

//TestManageHostBlock
func TestManageHostBlock(t *testing.T) {
	g := goblin.Goblin(t)
	c := Config{}
	c.Consul.Address = "localhost:8500"
	c.Consul.Scheme = "http"
	c.Vpn.Name = "tzk"
	g.Describe("ParseToFileFormat", func() {
		g.It("should insert or remplace the block in the host file",
			func(done goblin.Done) {
				handle := func(v *Vpn, close func()) {
					o := v.Hosts.manageHostBlock("")
					assert.True(g, strings.Contains(o, "node3.tzk.local"))
					assert.True(g, strings.Contains(o, "node4.tzk.local"))
					assert.True(g, strings.Contains(o, "node1.tzk.local"))
					assert.True(g, strings.Contains(o, "node2.tzk.local"))

					o = v.Hosts.manageHostBlock(`127.0.0.1 localhost
#/tzk/NoEdit
                    10.65.1.23	node20.tzk.local
                    10.65.1.206	node30.tzk.local
                    10.65.1.248	node40.tzk.local
                    #/tzk/NoEdit`)
					assert.True(g, strings.Contains(o, "node3.tzk.local"))
					assert.True(g, strings.Contains(o, "node4.tzk.local"))
					assert.True(g, strings.Contains(o, "node1.tzk.local"))
					assert.True(g, strings.Contains(o, "node2.tzk.local"))
					assert.False(g, strings.Contains(o, "node20.tzk.local"))
					assert.False(g, strings.Contains(o, "node30.tzk.local"))
					assert.False(g, strings.Contains(o, "node40.tzk.local"))

					close()
					done()

				}
				subnet1 := "10.65.1.0/24"
				bootstrapConsul(subnet1, c.Vpn.Name, c)
				DHCP(c, "node1")
				addHost(c, "node2", "pubkey2", "185.36.25.14")
				addHost(c, "node3", "pubkey3", "108.36.25.14")
				addHost(c, "node4", "pubkey4", "95.36.25.14")
				go WatchConsul(c, handle)

			})

	})
}

//TestFixName
func TestFixName(t *testing.T) {
	g := goblin.Goblin(t)
	g.Describe("fixName", func() {
		g.It("should replace non standard characters by _",
			func() {
				assert.Equal(g, fixName("ikkk?ikjkj-ljl{{"), "ikkk_ikjkj_ljl__")
			})

	})
}

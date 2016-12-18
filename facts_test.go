package main

import (
	"github.com/criloz/goblin"
	"github.com/stretchr/testify/assert"
	"testing"
)

//TestFacts_AddAddress check the specs for add address function
func TestFacts_AddAddress(t *testing.T) {
	g := goblin.Goblin(t)
	g.Describe("AddAddress", func() {
		g.It("Should add ipv4 or ipv6 addrs", func() {
			addrs := []string{"10.83.247.1",
				"2001:4860:4860::8888",
				"2001:4860:4860::8844",
				"192.168.0.11",
				"172.27.224.14"}
			f := Facts{}
			for _, addr := range addrs {
				f.AddAddress(addr)
			}
			assert.Len(g, f.Addresses, 5)
		})

		g.It("Should only accept global unicast addrs", func() {
			addrs := []string{"127.0.0.1",
				"fe80::5067:9ff:fe8a:d3c6",
				"fe80::7065:42ff:fe50:1dc4",
				"::1"}
			f := Facts{}
			for _, addr := range addrs {
				f.AddAddress(addr)
			}

			assert.Nil(g, f.Addresses)
		})
		g.It("Should not fail with bad addr", func() {
			addrs := []string{"127.0.0.1/24",
				"make_it_fail",
				"xxxxx.xxx.xxxx"}
			f := Facts{}
			for _, addr := range addrs {
				f.AddAddress(addr)
			}

		})

		g.It("Should not dupplicate the address", func() {
			addrs := []string{"172.27.224.14",
				"10.83.247.1",
				"10.83.247.1",
				"10.83.247.1",
				"2001:4860:4860::8888",
				"2001:4860:4860::8888",
				"2001:4860:4860::8844",
				"192.168.0.11",
				"172.27.224.14",
				"2001:4860:4860::8888"}
			f := Facts{}
			for _, addr := range addrs {
				f.AddAddress(addr)
			}
			assert.Len(g, f.Addresses, 5)

		})

	})
}

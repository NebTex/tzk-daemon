package main

import (
	log "github.com/Sirupsen/logrus"
	"net"
)

type Facts struct {
	Address       []string
	HasChanged    bool `json:"-"`
	Container     string
	City          string
	CountryCode   string
	CountryName   string
	RegionCode    string
	RegionName    string
	ZipCode       string
	TimeZone      string
	MetroCode     int
	Latitude      float32
	Longitude     float32
	ContinentCode string
	PublicKey     string
	Hostname      string `json:"-"`
}

//AddAddress add ipv4 or ipv6 address to the Fact map
//can only add global unicast address
func (f *Facts) AddAddress(addr string) {
	ip := net.ParseIP(addr)
	if ip == nil {
		log.Errorf("%s is not a valid Address\n", addr)
	}
	if ip.IsGlobalUnicast() {
		if !StringSliceContains(f.Address, addr) {
			f.Address = append(f.Address, addr)
			f.HasChanged = true
		}
	}
}

//StringSliceContains check if a string slice contains the searchString
func StringSliceContains(stringSlice []string, searchString string) bool {
	for _, value := range stringSlice {
		if value == searchString {
			return true
		}
	}
	return false
}

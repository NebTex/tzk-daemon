package commons

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io"
	"net/http"
	"os"
)

//GetGeoIP use the freegeoip.net api and gather the public ip
//of this node also other geo localization info
func (f *Facts) GetGeoIP() {

	resp, err := http.Get("https://freegeoip.net/json")
	if err != nil {
		log.Errorln(err)
		return
	}
	if resp.StatusCode == 403 {
		log.Errorln("freegeoip.net rate limit exceeded")
		return
	}
	if resp.StatusCode != 200 {
		_, err = io.Copy(os.Stderr, resp.Body)
		log.Errorln(err)

		return
	}
	var body struct {
		IP            string  `json:"ip"`
		City          string  `json:"city"`
		CountryCode   string  `json:"country_code"`
		CountryName   string  `json:"country_name"`
		RegionCode    string  `json:"region_code"`
		RegionName    string  `json:"region_name"`
		ZipCode       string  `json:"zip_code"`
		TimeZone      string  `json:"time_zone"`
		MetroCode     int     `json:"metro_code"`
		Latitude      float32 `json:"latitude"`
		Longitude     float32 `json:"longitude"`
		ContinentCode string
	}

	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		log.Errorln(err)
		return
	}
	if cc, ok := ContinentCodeLookup[body.CountryCode]; ok {
		body.ContinentCode = cc
	}

	f.City = body.City
	f.CountryCode = body.CountryCode
	f.CountryName = body.CountryName
	f.RegionCode = body.RegionCode
	f.RegionName = body.RegionName
	f.ZipCode = body.ZipCode
	f.TimeZone = body.TimeZone
	f.MetroCode = body.MetroCode
	f.Latitude = body.Latitude
	f.Longitude = body.Longitude
	f.ContinentCode = body.ContinentCode
	f.AddAddress(body.IP)
}

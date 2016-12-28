package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

func init() {
	log.SetFormatter(&log.TextFormatter{})
}

func initiate(c Config) *Host {
	initSubnet(c)
	h := Host{}
	h.Facts.GetContainerStatus()
	h.Facts.GetLocalAddresses()
	h.Facts.GetGeoIP()
	h.Facts.GetTincInfo(c, os.Hostname)
	h.Facts.SendToConsul(c)
	h.SetConfigConsul(c)
	DHCP(c, h.Facts.Hostname)
	handleConsulChange(c, h)
	return &h
}

func mainLoop(c Config, h *Host) {
	geoIPLimiter := 0

	for {
		h.Facts.GetContainerStatus()
		h.Facts.GetLocalAddresses()
		// get geo ip info each day
		if geoIPLimiter == 0 {
			h.Facts.GetGeoIP()
		}
		h.Facts.GetTincInfo(c, os.Hostname)
		h.Facts.SendToConsul(c)
		geoIPLimiter++
		if geoIPLimiter > 1440 {
			geoIPLimiter = 0
		}
		time.Sleep(60 * time.Second)
	}
}
func handleConsulChange(c Config, h Host) {
	var oldFiles *Files
	var subnet string
	go WatchConsul(c, func(v *Vpn, close func()) {
		if subnet != v.Subnet {
			DHCP(c, h.Facts.Hostname)
			subnet = v.Subnet
		}
		v.SetHostFile(h.Facts.Hostname)
		files := v.GenerateFiles(h.Facts.Hostname, c)
		if !files.Equal(oldFiles) {
			files.Write(c)
			oldFiles = files
			cmd := strings.Split(c.Vpn.ExecStart, " ")
			_, err := exec.Command(cmd[0], cmd[1:]...).Output()
			log.Error(err)
		}
	})
}

func main() {
	app := cli.NewApp()
	app.Description = "Gather some fact about the node and send it to the master"
	app.Version = "0.1"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "conf-dir",
			Value: "/etc/tzk",
			Usage: "config directory",
		},
	}

	app.Action = func(c *cli.Context) error {
		configFile := filepath.Join(c.String("conf-dir"), "tzk.toml")
		data, err := ioutil.ReadFile(configFile)
		if err != nil {
			log.Fatal(err)
		}
		config := Config{}
		if _, err := toml.Decode(string(data), &config); err != nil {
			log.Fatal(err)
		}
		mainLoop(config, initiate(config))
		return nil
	}

	err := app.Run(os.Args)
	checkFatal(err)

}

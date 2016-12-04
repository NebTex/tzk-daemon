package main

import (
	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	Vpn struct {
		Name string
	}
	Consul struct {
		Url      string
		ACLToken string
	}
}

func init() {
	log.SetFormatter(&log.TextFormatter{})
}

func main() {
	app := cli.NewApp()
	app.Description = "Gather some fact about the node and send it to the master"
	app.Version = "0.1"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "conf-dir",
			Value: "/etc/tzn",
			Usage: "config directory",
		},
	}

	app.Action = func(c *cli.Context) error {
		configFile := filepath.Join(c.String("conf-dir"), "tzn.toml")
		data, err := ioutil.ReadFile(configFile)
		if err != nil {
			log.Fatal(err)
		}
		config := Config{}
		//TODO: Add loop
		if _, err := toml.Decode(string(data), &config); err != nil {
			log.Fatal(err)
		}

		f := Facts{}
		geoIPLimiter := 0

		for {

			time.Sleep(60 * time.Second)
			f.GetContainerStatus()
			f.GetLocalAddresses()
			// get geo ip info each day
			if geoIPLimiter == 0 {
				f.GetGeoIP()
			}
			f.GetTincInfo(config)
			f.SendToConsul(config)
			geoIPLimiter += 1
			if geoIPLimiter > 1440 {
				geoIPLimiter = 0
			}
		}

		return nil

	}

	app.Run(os.Args)

}

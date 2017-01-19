package main

import (
    "io/ioutil"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"
    "tzk-daemon/commons"
    "github.com/BurntSushi/toml"
    log "github.com/Sirupsen/logrus"
    "github.com/urfave/cli"
    "fmt"
    "tzk-daemon/dhcp"
    "tzk-daemon/hosts"
    "net"
)

func init() {
    log.SetFormatter(&log.TextFormatter{})
}

func initiate(c commons.Config) *commons.Host {
    dhcp.InitSubnet(c)
    h := commons.Host{}
    h.Dumps = &commons.Dumps{}
    h.Facts.GetLocalAddresses()
    h.Facts.GetGeoIP()
    h.Facts.GetTincInfo(c, os.Hostname)
    h.Facts.SendToConsul(c)
    h.SetConfigConsul(c)
    ip, s := dhcp.DHCP(c, h.Facts.Hostname)
    h.VpnAddress = ip
    h.PodSubnet = s
    handleConsulChange(c, h)
    return &h
}

func mainLoop(c commons.Config, h *commons.Host) {
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
        h.Dumps.Get(c)
        h.SendDumpsToConsul(c)
    }
}

func handleConsulChange(c commons.Config, h commons.Host) {
    var oldFiles *Files
    var subnet string
    go hosts.WatchConsul(c, func(v *hosts.Vpn, close func()) {
        if subnet != v.Subnet {
            dhcp.DHCP(c, h.Facts.Hostname)
            subnet = v.Subnet
        }
        v.SetHostFile(h.Facts.Hostname)
        files := GenerateFiles(v, h.Facts.Hostname, c)
        if !files.Equal(oldFiles) {
            stopCMD := strings.Split(c.Vpn.ExecStop, " ")
            _, err := exec.Command(stopCMD[0], stopCMD[1:]...).Output()
            if err != nil {
                log.Error("Failed to stop tinc")
                log.Error(err)
            }
            files.Write(c)
            oldFiles = files
            startCMD := strings.Split(c.Vpn.ExecStart, " ")
            _, err = exec.Command(startCMD[0], startCMD[1:]...).Output()
            if err != nil {
                log.Error("Failed to start tinc")
                log.Error(err)
            }
        }
    })
}

func getConfig(context *cli.Context) commons.Config {
    configFile := filepath.Join("/etc/tzk.d/", "tzk.toml")
    data, err := ioutil.ReadFile(configFile)
    commons.CheckFatal(err)
    config := commons.Config{}
    if _, err := toml.Decode(string(data), &config); err != nil {
        log.Fatal(err)
    }
    return config
}
func getIP(config commons.Config) {
    h, err:=os.Hostname()
    commons.CheckFatal(err)
    ip, _ := dhcp.DHCP(config, h)
    fmt.Print(ip)
}

func main() {
    app := cli.NewApp()
    app.Description = "Gather some fact about the node and send it " +
        "to the master"
    app.Version = "0.2.0"
    
    app.Flags = []cli.Flag{
        cli.StringFlag{
            Name:  "conf-dir",
            Value: "/etc/tzk",
            Usage: "config directory",
        },
    }
    app.Commands = []cli.Command{
        {
            Name:    "run",
            Usage:   "run tzk daemon",
            Aliases: []string{},
            Action:  func(c *cli.Context) error {
                config := getConfig(c)
                mainLoop(config, initiate(config))
                return nil
            }}, {
            Name:        "get",
            Usage:       "get a resource",
            Subcommands: []cli.Command{
                {
                    Name:  "logs",
                    Usage: "show the logs",
                    Aliases: []string{},
                    Action: func(c *cli.Context) error {
                        logs := filepath.Join("/etc/tinc/tzk", "tinc.logs")
                        data, err := ioutil.ReadFile(logs)
                        commons.CheckFatal(err)
                        fmt.Print(string(data))
                        return nil
                    },
                },
                {
                    Name:  "podSubnet",
                    Usage: "podSubnet for this host",
                    Aliases: []string{},
                    Action: func(c *cli.Context) error {
                        log.SetLevel(log.ErrorLevel)
                        
                        config := getConfig(c)
                        h, err:=os.Hostname()
                        commons.CheckFatal(err)
                        _, PodSubnet := dhcp.DHCP(config, h)
                        ip, _, err := net.ParseCIDR(PodSubnet)
                        commons.CheckFatal(err)
                        ip[len(ip) - 1]++
                        fmt.Print(ip.String() + "/24")
                        return nil
                    }},
                {Name:  "ip",
                    Usage: "ip for this host",
                    Aliases: []string{},
                    Action: func(c *cli.Context) error {
                        log.SetLevel(log.ErrorLevel)
                        getIP(getConfig(c))
                        return nil
                    }}}}}
    err := app.Run(os.Args)
    commons.CheckFatal(err)
}





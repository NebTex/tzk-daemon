package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

//SetHostFile maintain updates the host file
func (v *Vpn) SetHostFile() {
	//read file
	d, err := ioutil.ReadFile("/etc/hosts")
	checkFatal(err)
	//replace
	o := v.Hosts.manageHostBlock(string(d))
	//save
	err = ioutil.WriteFile("/etc/hosts", []byte(o), 0644)
	checkFatal(err)
}

func (hs Hosts) manageHostBlock(input string) string {
	re := regexp.MustCompile(`(?ms)#\/tzn\/NoEdit(.)*#\/tzn\/NoEdit`)
	all := re.FindAllString(input, -1)
	if len(all) == 0 {
		return fmt.Sprintf(`%s
#/tzn/NoEdit
%s
#/tzn/NoEdit`, input, hs.parseToFileFormat())

	}
	return re.ReplaceAllString(input, fmt.Sprintf(`#/tzn/NoEdit
%s
#/tzn/NoEdit`, hs.parseToFileFormat()))
}

func (hs Hosts) parseToFileFormat() string {
	hFile := make([]string, len(hs))
	for _, h := range hs {
		entry := fmt.Sprintf("%s\t%s.%s.local",
			h.VpnAddress, h.Facts.Hostname, "tzn")
		hFile = append(hFile, entry)
	}
	return strings.Join(hFile, "\n")
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

func (f *Facts) SendToConsul(c Config) {
	if !f.HasChanged {
		return
	}

	b, err := json.Marshal(f)
	if err != nil {
		log.Errorln(err)
		return
	}

	url := fmt.Sprintf("%s/v1/kv/tzn/%s/facts", c.Consul.Url, f.Hostname)
	log.Infoln("Setting consul key ...")
	log.Infoln(url)

	//set new values
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(b))
	if err != nil {
		log.Errorln(err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Errorln("Failed to make a request to the consul service")
		log.Errorln(err)
		return
	}

	if resp.StatusCode != 200 {
		rb, _ := ioutil.ReadAll(resp.Body)
		log.Errorln(resp.StatusCode)
		log.Errorln(string(rb))
	}

	defer resp.Body.Close()
	f.HasChanged = false

}

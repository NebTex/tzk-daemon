package commons

import (
    "fmt"
    "strconv"
    
    log "github.com/Sirupsen/logrus"
    "github.com/hashicorp/consul/api"
)

func toByte(value interface{}) []byte {
    switch v := value.(type) {
    case string:
        return []byte(v)
    case int:
        return []byte(strconv.Itoa(v))
    case float32:
        return []byte(strconv.FormatFloat(float64(v), 'f', 4, 32))
    default:
    }
    return []byte("")
}

func (f *Facts) addKey(k string, v interface{}) *api.KVTxnOp {
    return &api.KVTxnOp{Verb: "set",
        Key:   fmt.Sprintf("tzk/Hosts/%s/Facts/%s", f.Hostname, k),
        Value: toByte(v)}
}

func (f *Facts) parseAddress(txn api.KVTxnOps) api.KVTxnOps {
    for address := range f.Addresses {
        txn = append(txn, f.addKey(fmt.Sprintf("Addresses/%s", address), ""))
    }
    return txn
}

//SendToConsul save facts in consul
func (f *Facts) SendToConsul(c Config) {
    if !f.HasChanged {
        return
    }
    
    ctx := api.KVTxnOps{f.addKey("Container", f.Container),
        f.addKey("City", f.City),
        f.addKey("CountryCode", f.CountryCode),
        f.addKey("RegionCode", f.RegionCode),
        f.addKey("RegionName", f.RegionName),
        f.addKey("ZipCode", f.ZipCode),
        f.addKey("TimeZone", f.TimeZone),
        f.addKey("MetroCode", f.MetroCode),
        f.addKey("Latitude", f.Latitude),
        f.addKey("Longitude", f.Longitude),
        f.addKey("ContinentCode", f.ContinentCode),
        f.addKey("PublicKey", f.PublicKey),
        f.addKey("HostName", f.Hostname)}
    ctx = f.parseAddress(ctx)
    log.Infoln("Storing info on Consul ...")
    //set new values
    client := GetConsulClient(c)
    ok, _, _, err := client.KV().Txn(ctx, nil)
    if err != nil || !ok {
        log.Errorln("Failed to make a request to the consul service")
        log.Fatal(err)
        return
    }
    f.HasChanged = false
}

func (h *Host) addDumpKey(k string, v string) *api.KVTxnOp {
    return &api.KVTxnOp{Verb: "set",
        Key:   fmt.Sprintf("tzk/Hosts/%s/Dumps/%s", h.Facts.Hostname, k),
        Value: []byte(v)}
}

//SendDumpsToConsul save dumps in consul
func (h *Host) SendDumpsToConsul(c Config) {
    
    ctx := api.KVTxnOps{h.addDumpKey("Nodes", h.Dumps.Nodes),
        h.addDumpKey("Connections", h.Dumps.Connections),
        h.addDumpKey("Subnets", h.Dumps.Subnets),
        h.addDumpKey("Edges", h.Dumps.Edges),
        h.addDumpKey("Graph", h.Dumps.Graph),
        h.addDumpKey("Invitations", h.Dumps.Invitations)}
    
    client := GetConsulClient(c)
    ok, _, _, err := client.KV().Txn(ctx, nil)
    if err != nil || !ok {
        log.Error("Dump: Failed to make a request to the consul service")
        log.Fatal(err)
        return
    }
}

## tzk-daemon

[![CircleCI](https://circleci.com/gh/NebTex/tzk-daemon.svg?style=svg)](https://circleci.com/gh/NebTex/tzk-daemon) [![Go Report Card](https://goreportcard.com/badge/github.com/NebTex/tzk-daemon)](https://goreportcard.com/report/github.com/NebTex/tzk-daemon) [![codecov](https://codecov.io/gh/NebTex/tzk-daemon/branch/master/graph/badge.svg)](https://codecov.io/gh/NebTex/tzk-daemon)

Small tool for coordinate mesh vpns (initially tinc) using consul as backend
 
 1. Continuously push node information to the consul backend
 
     - Interfaces ip
     - Geo-ip info (thanks to https://freegeoip.net)
     - Check if the node is a bare-metal or lxc container
     - Tinc public key of the node
 
 2. Automatically assign an ip to the node, if there is a subnet change, automatically  reassign a new ip to the node
 
 3. Maintain updated the `/etc/hosts` file, with a entry for each node, 
 with its respective vpn address 
 
 4. Write and maintain updated the tinc files, `tinc.conf`, `tinc-down`, `tinc-up` and the node-hosts files
 
### Vendoring 

trash was used  https://github.com/rancher/trash

### Licence

Copyright (c) 2016 NebTex

MIT 

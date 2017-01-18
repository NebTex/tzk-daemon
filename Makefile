
#==============================================================================
#											 			Linters
#==============================================================================

lint:
	gometalinter --disable-all --enable=dupl --enable=errcheck --enable=goconst \
	--enable=golint --enable=gosimple --enable=ineffassign --enable=interfacer \
	--enable=misspell --enable=staticcheck --enable=structcheck  \
	--enable=unused --enable=vet --enable=vetshadow --enable=lll \
	--line-length=80 --deadline=60s --vendor --dupl-threshold=500 ./...


test:
	go test -v tzk-daemon/commons
	go test -v tzk-daemon/dhcp
	go test -v tzk-daemon/hosts
	go test -v

.PHONY: lint

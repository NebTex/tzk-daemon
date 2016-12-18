
#==============================================================================
#											 			Linters
#==============================================================================

lint:
	gometalinter --disable-all --enable=dupl --enable=errcheck --enable=goconst \
	--enable=golint --enable=gosimple --enable=ineffassign --enable=interfacer \
	--enable=misspell --enable=staticcheck --enable=structcheck --enable=gocyclo \
	--enable=unused --enable=vet --enable=vetshadow --enable=lll \
	--line-length=80 --deadline=60s --vendor --dupl-threshold=100 ./...



.PHONY: lint

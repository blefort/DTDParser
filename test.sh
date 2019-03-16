#!/usr/bin/env bash
set -e
echo "" > coverage.txt
for d in $(go list ./... | grep -v vendor); do
    go test -overwrite=true -verbose=trace -coverprofile=profile.out -coverpkg=github.com/blefort/DTDParser/DTD,github.com/blefort/DTDParser/parser,github.com/blefort/DTDParser/scanner $d
    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done
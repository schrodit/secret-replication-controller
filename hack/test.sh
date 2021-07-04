#/bin/bash

eval "$(setup-envtest use -p env 1.21.x)"
go test ./... -coverprofile cover.out

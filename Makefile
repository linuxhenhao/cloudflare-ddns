ddns: *.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o cloudflare_ddns *.go
	chmod +x cloudflare_ddns

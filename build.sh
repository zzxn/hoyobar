gofmt -l -w -s . && \
goimports -l -w . && \
go build -o ./bin/hoyobar


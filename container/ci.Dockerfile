FROM golang:1.13.15-stretch

RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.41.1 -- -b /usr/local/bin 2>&1

FROM golang:1.17


WORKDIR /go/src/chrome-crawler
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

ENTRYPOINT ["chrome-crawler"]

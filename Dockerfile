FROM golang:1.17


WORKDIR /go/src/go-reflect
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

ENTRYPOINT ["go-reflect"]

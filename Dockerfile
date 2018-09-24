FROM golang:1.11
ADD . /go/src/inventory/
RUN go get github.com/gorilla/handlers
RUN go install /go/src/inventory
ENTRYPOINT /go/bin/inventory
EXPOSE 8081


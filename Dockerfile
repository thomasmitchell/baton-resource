FROM golang:1.15 AS build

COPY . /src
WORKDIR /src

RUN go get -d ./...
RUN mkdir /build
RUN go build -o /build/in    in/*.go
RUN go build -o /build/out   out/*.go 
RUN go build -o /build/check check/*.go


FROM ubuntu:bionic AS resource

RUN mkdir -p /opt/resource

COPY --from=build /build/in    /opt/resource/in
COPY --from=build /build/out   /opt/resource/out
COPY --from=build /build/check /opt/resource/check


FROM resource

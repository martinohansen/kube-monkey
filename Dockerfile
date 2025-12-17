########################
### Builder          ###
########################
FROM golang:1.24 AS builder
ARG TARGETOS
ARG TARGETARCH
RUN mkdir -p /kube-monkey
COPY ./ /kube-monkey/
WORKDIR /kube-monkey
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} make build

########################
### Final            ###
########################
FROM scratch
COPY --from=builder /kube-monkey/kube-monkey /kube-monkey
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
ENTRYPOINT ["/kube-monkey"]

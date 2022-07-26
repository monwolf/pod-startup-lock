FROM golang:1.18-bullseye as builder
ENV CGO_ENABLED=0
ENV GOOS=linux
WORKDIR /go/src/github.com/monwolf/pod-startup-lock/
COPY . .
RUN cd init &&	go build -a -o bin/init && cd ..
RUN cd k8s-health && go test -cover -v ./... &&	go build -a -o bin/k8s-health && cd ..
RUN cd lock && go test -cover -v ./... &&	go build -a -o bin/lock && cd ..
RUN cd hashi-health && go test -cover -v ./... &&	go build -a -o bin/hashi-health && cd ..



FROM scratch as init
WORKDIR /
COPY --from=builder /go/src/github.com/monwolf/pod-startup-lock/init/bin/init /init
ENTRYPOINT ["./init"]

FROM scratch as k8s-health
WORKDIR /
COPY --from=builder /go/src/github.com/monwolf/pod-startup-lock/k8s-health/bin/k8s-health /k8s-health
ENTRYPOINT ["./k8s-health"]

FROM scratch as lock
WORKDIR /
COPY --from=builder /go/src/github.com/monwolf/pod-startup-lock/lock/bin/lock /lock
ENTRYPOINT ["./lock"]

FROM scratch as hashi-health
WORKDIR /
COPY --from=builder /go/src/github.com/monwolf/pod-startup-lock/hashi-health/bin/hashi-health /hashi-health
ENTRYPOINT ["./hashi-health"]

FROM golang:1.18-bullseye as builder
WORKDIR /go/src/github.com/monwolf/pod-startup-lock/
COPY . .
RUN cd init &&	go build -a -o bin/init && cd ..
RUN cd k8s-health && go test -cover -v ./... &&	go build -a -o bin/k8s-health && cd ..
RUN cd lock && go test -cover -v ./... &&	go build -a -o bin/lock && cd ..

FROM scratch as init
COPY --from=builder /go/src/github.com/monwolf/pod-startup-lock/init/bin/init init
ENTRYPOINT ["./init"]

FROM scratch as k8s-health
COPY --from=builder /go/src/github.com/monwolf/pod-startup-lock/k8s-health/bin/k8s-health lock
ENTRYPOINT ["./lock"]

FROM scratch as lock
COPY --from=builder /go/src/github.com/monwolf/pod-startup-lock/lock/bin/lock lock
ENTRYPOINT ["./lock"]
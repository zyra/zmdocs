FROM golang:1.12-alpine as builder
RUN apk add git make
WORKDIR /root/wd
COPY . .
RUN make -j$(nproc)

FROM alpine
COPY --from=builder /root/wd/zmdocs_linux_amd64 /usr/bin/zmdocs
RUN chmod +x /usr/bin/zmdocs

ENTRYPOINT ["zmdocs"]
FROM golang:latest AS builder
WORKDIR dump
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o solrdump .
RUN chmod +x ./runner.sh
ENTRYPOINT ["/bin/bash","./runner.sh"]

#FROM ubuntu
#WORKDIR app
#COPY --from=builder /dump/solrdump .
#COPY runner.sh .
#RUN chmod +x ./runner.sh
#ENTRYPOINT ["/bin/bash","/app/runner.sh"]





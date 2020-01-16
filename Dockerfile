# build stage
FROM golang AS build-env
WORKDIR /src/
ADD net-report.go /src/
ADD go.mod /src/
RUN cd /src && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o net-report_linux64 net-report.go

# final stage
FROM google/cloud-sdk
WORKDIR /app/
COPY --from=build-env /src/net-report_linux64 /usr/bin/
ENTRYPOINT /bin/bash
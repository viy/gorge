##############################
# Base dev image             #
##############################

FROM golang:1.13.7-buster as development

ENV GO111MODULE=on

RUN apt-get update && \
    # Install Proj - C library for coordinate system conversion and its requirements 
    apt-get install -y libproj-dev \
    # Graphviz is needed for pprof
    graphviz 

WORKDIR /workspace

################################
# Test/lint production   image #
################################

FROM development as tester

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN make test && \
    make lint && \
    make typescript

################################
# Builder for production image #
################################

FROM tester as builder

ARG VERSION=0.0.0

RUN make build

################################
# Production image             #
################################
FROM gcr.io/distroless/base-debian10 as production

COPY --from=builder /go/bin/gorge-server /go/bin/gorge-cli /usr/local/bin/

EXPOSE 7080

ENTRYPOINT ["gorge-server"]
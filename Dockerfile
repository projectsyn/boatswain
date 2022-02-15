# Authors:
# Simon Gerber <simon.gerber@vshn.ch>
#
# License:
# Copyright (c) 2020, VSHN AG, <info@vshn.ch>
# Licensed under "BSD 3-Clause". See LICENSE file.

#####################
# STEP 1 build binary
#####################
FROM golang:1.17 as builder

ARG BINARY_VERSION

# Workdir must be outside of GOPATH because of go mod usage
WORKDIR /src/boatswain

# Download modules for leveraging docker build cache
COPY go.mod go.sum ./
RUN go mod download

# Add code
COPY . .

# Run tests and build Signalilo
RUN make test
RUN make build

############################
# STEP 2 build runtime image
############################
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /src/boatswain/boatswain /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/boatswain"]

FROM node:16 AS webbuilder

WORKDIR /web
COPY ./VERSION .
COPY ./web/default ./default

RUN npm install --prefix /web/default && \
    DISABLE_ESLINT_PLUGIN='true' CI=false \
    REACT_APP_VERSION=$(cat ./VERSION) npm run build --prefix /web/default

FROM golang:alpine AS gobuilder

RUN apk add --no-cache gcc musl-dev sqlite-dev build-base

ENV GO111MODULE=on CGO_ENABLED=1 GOOS=linux

WORKDIR /build
ADD go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=webbuilder /web/build ./web/build

RUN go build -trimpath -ldflags "-s -w -X 'github.com/jhopan/Panway/common.Version=$(cat VERSION)' -linkmode external -extldflags '-static'" -o panway

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

COPY --from=gobuilder /build/panway /panway

EXPOSE 3000
WORKDIR /data
VOLUME ["/data"]
ENTRYPOINT ["/panway"]
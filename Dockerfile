FROM --platform=linux/amd64 golang:alpine AS build

RUN apk add git sqlite sqlite-dev build-base build-base

WORKDIR /src/

COPY go.* /src/

RUN go mod download -x

COPY . /src

#Compiler Settings
ENV CGO_ENABLED=1

# for full parings check out https://go.dev/doc/install/source#environment
ENV GOOS=linux
# this will be the target cpu arch
# Can be amd64 arm64 386 ppc64
ENV GOARCH=amd64

RUN go build -o /out/app -tags "fts5" .

# if you need certificates use: alpine
# otherwise just use: scratch
FROM alpine AS run

RUN apk add build-base

COPY --from=build /out/app /

ENV MARK_STORE_LOCATION=/data
# if needed
EXPOSE 1995

ENTRYPOINT [ "/app", "search"]

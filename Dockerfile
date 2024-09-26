FROM golang:1.22-alpine as base

RUN apk update \
  && apk add --no-cache bash~=5 \
  && apk add --no-cache make~=4 \
  && apk add --no-cache build-base~=0.5 \
  && apk add --no-cache gettext~=0.22 \
  && apk add --no-cache --update-cache --upgrade curl~=8 \
  && apk add --no-cache git~=2

ARG PROJECT_ROOT
ENV CGO_ENABLED=1 \
  GOROOT='/usr/local/go' \
  GO111MODULE='on' \
  PROJECT_ROOT=${PROJECT_ROOT} \
  DEBUG_DLV="0" \
  APP_ENV='prod'

WORKDIR ${PROJECT_ROOT}

COPY go.sum .
COPY go.mod .

RUN go mod download
RUN go mod tidy

COPY . .

##### development ##############################################################
FROM base as development

# hadolint ignore=DL4006
RUN curl -sSfL "https://raw.githubusercontent.com/cosmtrek/air/master/install.sh" | sh -s -- -b "$(go env GOPATH)/bin"
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

ENV PATH="/go/bin:${PATH}"

CMD ["/bin/sh", "-c", "air"]

##### build ####################################################################
FROM base as build
RUN bash script/build.sh

##### production ###############################################################
FROM golang:1.22-alpine AS production

RUN apk update \
  && apk add --no-cache bash~=5 \
  && apk add --no-cache curl~=8

ARG PROJECT_ROOT
ENV APP_ENV='prod' \
  PATH="/go/bin:${PATH}"

RUN adduser -D -s /bin/sh -u 241 app

COPY --from=build "${PROJECT_ROOT}/target/whalebone-clients" "/bin/whalebone-clients"

RUN go install github.com/pressly/goose/v3/cmd/goose@latest
USER app

CMD ["/bin/sh", "-c", "/bin/whalebone-clients"]

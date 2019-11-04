FROM golang:1.13.4-alpine3.10 as builder

ARG GIT_TAG=master

RUN echo GIT_TAG=${GIT_TAG}

WORKDIR /opt/app

RUN apk add --update nodejs npm yarn bash

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN cd spa && \
     npm install react-scripts -g --silent && \
     cd ..

RUN go generate
RUN CGO_ENABLED=0 GOOS=linux go build

######## 
FROM alpine:latest

COPY --from=builder /opt/app/vault-token-issuer /usr/local/bin/vault-token-issuer

CMD ["vault-token-issuer"] 
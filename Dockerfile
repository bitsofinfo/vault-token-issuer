FROM golang:1.13.4

RUN go get github.com/bitsofinfo/vault-token-issuer

CMD vault-token-issuer
FROM golang:latest

ENV GOPATH /var/lib/go
ENV SRCPATH $GOPATH/src/github.com/SArtemJ/CurrencyGameExample/steam
ENV PATH $PATH:$GOPATH/bin:/go/bin

RUN go get -u github.com/Masterminds/glide

COPY glide.yaml glide.lock $SRCPATH/
WORKDIR $SRCPATH
RUN glide install
COPY / $SRCPATH

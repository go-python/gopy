FROM golang:onbuild
RUN apt-get update && apt-get install -y pkg-config python2.7-dev && apt-get clean
CMD /go/bin/app

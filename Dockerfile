FROM alpine:latest

LABEL maintainer="Petr Kotas<petr.kotas@gmail.com>"

RUN mkdir -p /usr/app
WORKDIR /usr/app

ADD ./lockvalidation /usr/app/lockvalidation 

CMD ["./lockvalidation", "--tlsCertFile", "/etc/lockvalidation/cert/cert.pem", "--tlsKeyFile", "/etc/lockvalidation/cert/key.pem", "-v", "1"]

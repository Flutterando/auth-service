FROM golang:1.12.7-buster
EXPOSE 8081
RUN mkdir app
COPY . /app
WORKDIR /app
RUN go install
ENTRYPOINT ["auth_server"]
FROM golang:1.18.5

WORKDIR /go/src
ENV PATH="/go/bin:${PATH}"
ENV GO111MODULE=on
ENV CGO_ENABLED=1

RUN apt-get update

RUN go mod init github.com/junimslage10/go-finance

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.15.2 && \ 
    go get github.com/lib/pq@v1.10.0 && \
    go get github.com/stretchr/testify@v1.7.0 && \
    go get -u github.com/gin-gonic/gin@v1.8.1 && \
    go get github.com/joho/godotenv@v1.4.0 && \
    go get -u github.com/golang-jwt/jwt/v4 && \
    go install github.com/cosmtrek/air@v1.40.4
    
# RUN go install github.com/kyleconroy/sqlc/cmd/sqlc@v1.15.0

CMD ["tail", "-f", "/dev/null"]

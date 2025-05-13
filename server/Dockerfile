FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

RUN go install github.com/beego/bee/v2@latest

COPY . .

RUN mkdir -p ./logger/logs

EXPOSE 8080

CMD ["bee", "run"] 
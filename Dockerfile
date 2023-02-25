FROM golang:1.20-alpine as build

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

RUN go build -v -o ./bot ./cmd/student-bot/main.go


FROM alpine:3.17 as release

COPY --from=build /app/bot /app/bot

CMD ["/app/bot"]

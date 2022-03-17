FROM --platform=linux/amd64 golang

RUN mkdir /app
WORKDIR /app

COPY . .


ENV API_URL=https://www.mercadobitcoin.net/api

RUN go get
RUN go build -ldflags="-w -s" -o main

EXPOSE 80

CMD ["./main"]

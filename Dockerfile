FROM golang

RUN mkdir /app
WORKDIR /app

COPY . .

ENV API_URL=https://www.mercadobitcoin.net/api
ENV LISTEN_EVENTS=$LISTEN_EVENTS

RUN go get
RUN go build -o main

EXPOSE 80

CMD ["./main"]

FROM golang

WORKDIR /app

RUN go build -o app.exe ./cmd

COPY app.exe .
COPY config .

CMD [ "./app" ]
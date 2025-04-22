FROM golang

WORKDIR /app

COPY app.exe .
COPY config .

CMD [ "./app" ]
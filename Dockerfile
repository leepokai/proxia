FROM golang:1.22
WORKDIR /app
COPY . .
RUN go build -o gateway .
EXPOSE 8080
CMD ["./gateway"]



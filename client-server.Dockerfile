FROM golang:1.24.5

WORKDIR /root

# Copy everything from your project into the image
COPY . .
RUN mkdir ./bin

RUN go mod tidy
RUN go build -o bin/ github.com/jackdelahunt/protoexplore/server
RUN go build -o bin/ github.com/jackdelahunt/protoexplore/cli

# Run the server when the container starts
CMD ["./bin/server"]
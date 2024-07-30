# Dockerfile
FROM golang:1.21

# Set the Current Working Directory inside the container
WORKDIR /metastore

# Copy go mod and sum files
COPY go.mod ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source(application code) from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o exe .

# Expose port 7070 to the outside world
EXPOSE 7070

# Command to run the executable
CMD ["./exe"]

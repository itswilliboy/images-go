FROM golang:alpine

WORKDIR /app

COPY . .

RUN --mount=type=cache,target=/cache go mod download

EXPOSE 3000

ENTRYPOINT [ "go", "run", "." ]
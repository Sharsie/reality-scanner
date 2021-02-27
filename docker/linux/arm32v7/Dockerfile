FROM arm32v7/alpine:3.12 as alpine

# Install ca certificates to enable https requests
RUN apk --no-cache add ca-certificates


FROM arm32v7/golang:1.15.5 as build

ARG APP_NAME
WORKDIR /go/src/github.com/Sharsie/$APP_NAME/

COPY ["./cmd", "./cmd"]
COPY ["./go.mod", "./go.mod"]
COPY ["./go.sum", "./go.sum"]

ENV GOOS=linux
ENV GOARCH=arm
ENV GOARM=7
ARG COMMAND
ARG VERSION
ARG COMMIT_HASH
ARG REPO_URL

RUN CGO_ENABLED=0 go build -mod readonly -ldflags "-s -w \
    -X github.com/Sharsie/$APP_NAME/cmd/$COMMAND/version.Tag=$VERSION \
    -X github.com/Sharsie/$APP_NAME/cmd/$COMMAND/version.Commit=$COMMIT_HASH \
    -X github.com/Sharsie/$APP_NAME/cmd/$COMMAND/version.buildTime=$BUILD_TIME \
    -X github.com/Sharsie/$APP_NAME/cmd/$COMMAND/version.repoUrl=$REPO_URL" -o ./bin/$COMMAND ./cmd/$COMMAND


FROM scratch

COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ARG COMMAND
ARG APP_NAME

COPY --from=build /go/src/github.com/Sharsie/$APP_NAME/bin/$COMMAND /app

ENTRYPOINT ["/app"]

FROM golang:alpine AS build
ARG COMMIT_SHA
ARG TARGETOS
WORKDIR /build
COPY . .
RUN echo $TARGETOS
RUN sh build.sh $COMMIT_SHA

FROM alpine
WORKDIR /app
COPY --from=build /build/output .
ENTRYPOINT ["./bootstrap.sh"]

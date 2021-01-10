################################################ Go Build Stage

FROM golang:1.15.5 as build

ENV USER=ocuser USERID=11235813
RUN adduser --disabled-password --gecos "" --home "/nonexistent" \
    --shell "/sbin/nologin" --no-create-home --uid "${USERID}" "${USER}"

RUN go get github.com/orcunuso/ocinfo
WORKDIR /go/src/github.com/orcunuso/ocinfo
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/ocinfo

############################################# Image Build Stage

FROM scratch
ARG APPLICATION="ocinfo"
ARG REFNAME="orcunuso/ocinfo"

LABEL org.opencontainers.image.ref.name="${REFNAME}" \
      org.opencontainers.image.authors="Ozan Orcunus <@orcunuso>" \
      org.opencontainers.image.documentation="https://github.com/${REFNAME}/README.md" \
      org.opencontainers.image.title="OCinfo" \
      org.opencontainers.image.description="Reporting tool for multiple OpenShift clusters" \
      org.opencontainers.image.licenses="Apache 2.0" \
      org.opencontainers.image.source="https://github.com/${REFNAME}"

COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group
COPY --from=build /go/bin/${APPLICATION} /${APPLICATION}

USER ocuser
ENTRYPOINT ["/ocinfo"]
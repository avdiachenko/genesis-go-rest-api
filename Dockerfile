FROM golang:1.19.0

WORKDIR /usr/src/app

ARG UID=1000
ARG GID=1000

RUN groupadd -g "${GID}" defaultuser \
  && useradd --create-home --no-log-init -u "${UID}" -g "${GID}" defaultuser

USER defaultuser

COPY . .
RUN go mod tidy

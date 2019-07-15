FROM jrottenberg/ffmpeg

RUN apt-get update

COPY . /app
WORKDIR /app

EXPOSE 1234

ENTRYPOINT ./bin/server


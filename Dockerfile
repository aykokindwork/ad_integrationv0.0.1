FROM ubuntu:latest
LABEL authors="aykokind"

ENTRYPOINT ["top", "-b"]
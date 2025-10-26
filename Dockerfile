FROM ubuntu:latest
LABEL authors="whoami"

ENTRYPOINT ["top", "-b"]
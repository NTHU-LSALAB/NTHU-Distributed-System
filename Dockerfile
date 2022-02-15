FROM gcr.io/distroless/base-debian11 AS base

COPY bin/app/cmd /cmd

CMD ["/cmd"]

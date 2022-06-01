FROM gcr.io/distroless/base-debian11 AS base

COPY bin/app/cmd /cmd
COPY bin/app/static /static

CMD ["/cmd"]

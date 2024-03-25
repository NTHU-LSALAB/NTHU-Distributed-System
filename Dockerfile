FROM gcr.io/distroless/base-debian12 AS base

COPY bin/app/cmd /cmd
COPY bin/app/static /static

CMD ["/cmd"]

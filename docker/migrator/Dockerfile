# Dockerfile
FROM arigaio/atlas:latest AS atlas

FROM alpine:latest

RUN apk add --no-cache postgresql postgresql-client bash

WORKDIR /app

COPY --from=atlas /atlas ./atlas
COPY migrations ./migrations
COPY ./docker/migrator/entrypoint.sh ./entrypoint.sh

RUN addgroup -S migrator && adduser -S migrator -G migrator

RUN chmod -R 765 . \
	&& chown -R migrator:migrator .

USER migrator

ENTRYPOINT ["/app/entrypoint.sh"]

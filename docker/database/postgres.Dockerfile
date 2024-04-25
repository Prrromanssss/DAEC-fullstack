FROM postgres:16-alpine3.19

ENV POSTGRES_PASSWORD=postgres

ENV POSTGRES_DB=daec

COPY ./sql/daec.sql /docker-entrypoint-initdb.d/
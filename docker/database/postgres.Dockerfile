FROM postgres:16-alpine3.19

ENV POSTGRES_PASSWORD=postgres

ENV POSTGRES_DB=daec

COPY ./backend/sql/daec.sql /docker-entrypoint-initdb.d/
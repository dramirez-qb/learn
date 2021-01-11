FROM node:12-alpine

ARG PROXY=

RUN mkdir -p /app \
    && chown -R 1001 /app

USER 1001

ENV PORT 3000

EXPOSE 3000

CMD ["main.js"]

WORKDIR /app

COPY ./main.js ./main.js

COPY ./package.json ./package.json

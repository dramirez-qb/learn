FROM maven:3.6.3-jdk-8-slim as builder
LABEL maintainer="Daniel Ramirez"

RUN mkdir /build
ADD . /build/
WORKDIR /build

COPY src /build
COPY pom.xml /build

RUN mvn clean package


FROM openjdk:8-jre-alpine as prod

ARG PROXY=
RUN mkdir -p /app \
    && chown -R 1001 /app
USER 1001
WORKDIR /app
COPY --from=builder /build/target/*.jar /app/app.jar
EXPOSE 8080
ENTRYPOINT [ "java" ]
CMD [ "-jar", "/app/app.jar" ]

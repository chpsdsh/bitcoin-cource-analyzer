FROM eclipse-temurin:17-jre-alpine

ARG JMETER_VERSION=5.6.3
ARG JMETER_BASE_URL=https://dlcdn.apache.org/jmeter/binaries

RUN apk add --no-cache bash curl tar

WORKDIR /opt

RUN curl -fsSL "${JMETER_BASE_URL}/apache-jmeter-${JMETER_VERSION}.tgz" -o /tmp/jmeter.tgz \
    && tar -xzf /tmp/jmeter.tgz -C /opt \
    && rm /tmp/jmeter.tgz \
    && ln -s "/opt/apache-jmeter-${JMETER_VERSION}" /opt/jmeter

ENV JMETER_HOME=/opt/jmeter
ENV PATH="${JMETER_HOME}/bin:${PATH}"

ENTRYPOINT ["jmeter"]

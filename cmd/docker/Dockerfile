FROM chromedp/headless-shell:latest

ARG BUILD_HOME
ARG APP_HOME
ARG APP
ARG DOCKER_UID
ARG DOCKER_USER
ARG DOCKER_GID
ARG DOCKER_GROUP
ARG DOCKER_TZ
ARG HEALTHCHECK

ENV BUILD_HOME=${BUILD_HOME}
ENV APP_HOME=${APP_HOME}
ENV APP=${APP}
ENV DOCKER_UID=${DOCKER_UID}
ENV DOCKER_USER=${DOCKER_USER}
ENV DOCKER_GID=${DOCKER_GID}
ENV DOCKER_GROUP=${DOCKER_GROUP}
ENV DOCKER_TZ=${DOCKER_TZ}
ENV HEALTHCHECK=${HEALTHCHECK}

USER root

COPY ${BUILD_HOME}/ ${APP_HOME}/

RUN \
  apt update -y && \
  apt upgrade -y && \
  apt clean && \
  if [ -n "${DOCKER_TZ}" ]; then \
    rm -f /etc/localtime && \
    ln -s /usr/share/zoneinfo/${DOCKER_TZ} /etc/localtime; \
  fi && \
  addgroup --gid ${DOCKER_GID} ${DOCKER_GROUP} && \
  adduser --home ${APP_HOME} --gecos ${DOCKER_USER} --uid ${DOCKER_UID} --ingroup ${DOCKER_GROUP} --shell /bin/sh --disabled-password ${DOCKER_USER} && \
  chown -R ${DOCKER_USER}:${DOCKER_GROUP} ${APP_HOME} && \
  rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* \
  ;

USER ${DOCKER_USER}

WORKDIR ${APP_HOME}

ENTRYPOINT [ "./balance-collector", "--config", "config/balance-collector.toml" ]

HEALTHCHECK CMD bash -c "${HEALTHCHECK}"

FROM chromedp/headless-shell:latest

ARG BUILD_HOME
ARG APP_HOME
ARG APP
ARG DOCKER_UID
ARG DOCKER_USER
ARG DOCKER_GID
ARG DOCKER_GROUP
ARG DOCKER_TZ

COPY ${BUILD_HOME}/ ${APP_HOME}/

RUN \
  apt update -y && \
  apt upgrade -y && \
  apt install -y curl && \
  apt clean && \
  if [ -n "${DOCKER_TZ}" ]; then \
    rm -f /etc/localtime && \
    ln -s /usr/share/zoneinfo/${DOCKER_TZ} /etc/localtime; \
  fi && \
  addgroup --gid ${DOCKER_GID} ${DOCKER_GROUP} && \
  adduser --home ${APP_HOME} --gecos ${DOCKER_USER} --uid ${DOCKER_UID} --ingroup ${DOCKER_GROUP} --shell /bin/sh --disabled-password ${DOCKER_USER} && \
  chown -R ${DOCKER_USER}:${DOCKER_GROUP} ${APP_HOME} && \
  rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

USER ${DOCKER_USER}

WORKDIR ${APP_HOME}

ENTRYPOINT [ "./balance-collector", "--config", "config/balance-collector.toml" ]

HEALTHCHECK CMD bash -c ". .env && curl -fs http://localhost:\${PORT}/ || exit 1"

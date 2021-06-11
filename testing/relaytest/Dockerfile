FROM ubuntu:20.04

# Install curl
#
RUN apt-get update                                          \
  && apt-get -y --force-yes install --no-install-recommends \
    curl                                                    \
  && apt-get clean                                          \
  && apt-get autoclean                                      \
  && apt-get autoremove                                     \
  && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

ADD splitecho.sh /opt/app/splitecho.sh

ENTRYPOINT [ "/bin/bash", "/opt/app/splitecho.sh" ]

CMD [ "foo" ]

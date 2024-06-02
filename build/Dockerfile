FROM alpine:3.20

ENV OPERATOR=/usr/local/bin/kubedge-arpscan-operator \
    USER_UID=1001 \
    USER_NAME=kubedge-arpscan-operator

# install operator binary
COPY build/_output/bin/kubedge-arpscan-operator ${OPERATOR}

COPY build/bin /usr/local/bin

RUN  /usr/local/bin/user_setup

COPY build/arpscan-templates /opt/kubedge-operators/arpscan-templates

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}

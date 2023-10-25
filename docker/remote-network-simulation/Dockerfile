FROM 0kn/trellis:latest

# install ssh server
RUN apt update \
  && apt install --no-install-recommends -y -q \
    openssh-server \
  && rm -rf /var/lib/apt/lists/*

# sshd will fail without this
RUN mkdir /var/run/sshd

# add an ssh user, used for config share by coordinator to servers
ARG SSHUSER=ec2-user
RUN useradd -m -U -s /bin/bash ${SSHUSER}

# set permissions for user to exeucte (files are generated there)
RUN chown -R ${SSHUSER}:${SSHUSER} /src

# avoid base image rebuild for faster dev
COPY --chown=${SSHUSER}:${SSHUSER} *.sh /tmp/

# setup trellis expectations for config and binary locations
RUN mkdir -p /home/${SSHUSER}/go/bin \
  && ln -s \
    /src/go/trellis/cmd/server/server \
   /home/${SSHUSER}/go/bin/ \
  && chown -R ${SSHUSER}:${SSHUSER} /home/${SSHUSER}/go

EXPOSE 22
CMD ["/usr/sbin/sshd", "-D"]

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

# set permissions for user to execute (files are generated there)
RUN chown -R ${SSHUSER}:${SSHUSER} /src

EXPOSE 22
CMD ["/usr/sbin/sshd", "-D"]

FROM centos:8

RUN dnf install -y gdb && \
    dnf clean all && \
    rm -rf /var/cache/dnf

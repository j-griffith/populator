FROM fedora:29
LABEL maintainer="John Griffith <john.griffith8@gmail.com>"
RUN dnf update -y && dnf install -y git && dnf clean packages
COPY populate.bash /usr/local/bin/
RUN ln -s usr/local/bin/populate.bash
ENTRYPOINT ["populate.bash"]

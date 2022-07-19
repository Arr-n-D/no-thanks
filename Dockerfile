FROM golang:1.18 as build
WORKDIR /usr/no-thanks/
COPY . ./
RUN go mod download

COPY . .
WORKDIR /usr/no-thanks/browsers
RUN apt-get update && apt-get upgrade && apt-get install unzip
# RUN export PATH
# RUN go run init.go --alsologtostderr
WORKDIR /usr/no-thanks/
RUN go build -o ./no-thanks

FROM golang:1.18 as release
WORKDIR /root/
RUN apt-get update && apt-get upgrade && apt-get install -y xvfb && apt-get install -y openjdk-11-jdk ca-certificates-java && \
    apt-get clean && \
    update-ca-certificates -f
RUN  apt-get update \
    && apt-get install -y wget gnupg ca-certificates \
    && wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add - \
    && sh -c 'echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google.list' \
    && apt-get update \
    # We install Chrome to get all the OS level dependencies, but Chrome itself
    # is not actually used as it's packaged in the node puppeteer library.
    # Alternatively, we could could include the entire dep list ourselves
    # (https://github.com/puppeteer/puppeteer/blob/master/docs/troubleshooting.md#chrome-headless-doesnt-launch-on-unix)
    # but that seems too easy to get out of date.
    && apt-get install -y google-chrome-stable \
    && rm -rf /var/lib/apt/lists/* \
    && wget --quiet https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh -O /usr/sbin/wait-for-it.sh \
    && chmod +x /usr/sbin/wait-for-it.sh
ENV JAVA_HOME /usr/lib/jvm/java-11-openjdk-amd64/
ENV PATH $PATH:$JAVA_HOME/bin
RUN export JAVA_HOME
RUN echo google-chrome --version
COPY --from=build /usr/no-thanks ./
ENTRYPOINT [ "./no-thanks" ]
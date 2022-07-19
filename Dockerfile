FROM golang:1.18 as build
WORKDIR /usr/no-thanks/
COPY . ./
RUN go mod download

COPY . .
WORKDIR /usr/no-thanks/browsers
RUN apt-get update && apt-get upgrade && apt-get install unzip
RUN export PATH
RUN go run init.go --alsologtostderr  --download_browsers --download_latest
WORKDIR /usr/no-thanks/
RUN go build -o ./no-thanks

FROM golang:1.18 as release
WORKDIR /root/
RUN apt-get update && apt-get upgrade && apt-get install -y xvfb && apt-get install -y openjdk-11-jdk ca-certificates-java && \
    apt-get clean && \
    update-ca-certificates -f
ENV JAVA_HOME /usr/lib/jvm/java-11-openjdk-amd64/
ENV PATH $PATH:$JAVA_HOME/bin
RUN export JAVA_HOME
COPY --from=build /usr/no-thanks ./
ENTRYPOINT [ "./no-thanks" ]
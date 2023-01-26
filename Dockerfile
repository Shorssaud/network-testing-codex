# This is the image for v3


FROM ubuntu:22.04

RUN apt update -qq && \
    DEBIAN_FRONTEND="noninteractive" apt install -yq cmake curl make xz-utils 
RUN curl https://nim-lang.org/choosenim/init.sh -sSf | sh -s -- -y

ARG corecount=1
COPY . .
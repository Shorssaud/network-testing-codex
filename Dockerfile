FROM nimlang/nim as builder

RUN apt update -qq && \
    DEBIAN_FRONTEND="noninteractive" apt install -yq cmake curl make

FROM builder
ARG corecount=1
COPY nim-codex .

# TODO add a processor count variable argument
# RUN make -j6 update
# RUN make -j6 USE_SYSTEM_NIM=1 exec

RUN make -j${corecount} update

RUN make -j${corecount} USE_SYSTEM_NIM=1 exec

ENTRYPOINT ["."]

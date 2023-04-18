FROM debian:stable-slim

ARG DETECTLANG_APIKEY

ARG HOME=/root
ARG GOPATH=${HOME}/go
ARG CHLOE_DIR=${GOPATH}/src/chloe
ARG GO=${GOPATH}/bin/go

ENV GOPATH=${GOPATH}

WORKDIR ${CHLOE_DIR}

RUN apt update
RUN apt install -y git protobuf-compiler python3 python3-pip ffmpeg wget

RUN cd ${HOME} && \
      arch="$(uname -m)"; \
      case "$arch" in \
          x86_64) export GOARCH='amd64' ;; \
          aarch64) export GOARCH='armv6l' ;; \
      esac; \
      wget -c -t 0 https://go.dev/dl/go1.20.3.linux-${GOARCH}.tar.gz
RUN cd ${HOME} && tar xzvf go*.tar.gz && rm -f go*.tar.gz 

RUN git clone https://github.com/DiamondGo/Chloe.git ${CHLOE_DIR}
RUN cd ${CHLOE_DIR} && git submodule update --init --recursive

RUN cd ${CHLOE_DIR}/pyservice && pip install -r requirements.txt

RUN cd ${CHLOE_DIR}/pyservice && python3 -m grpc_tools.protoc -I./proto --python_out=pb --grpc_python_out=pb proto/tts.proto
RUN cd ${CHLOE_DIR} && ${GO} install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
    && ${GO} install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest \
    && export PATH=${GOPATH}/bin:$PATH \
    && protoc --go_out=. --go-grpc_out=. pyservice/proto/tts.proto
RUN cd ${CHLOE_DIR} && ${GO} build
RUN mkdir -p ${CHLOE_DIR}/log

RUN echo "cd ${CHLOE_DIR}/pyservice && python3 main.py ${DETECTLANG_APIKEY} pyservice &" >> ${HOME}/start.sh
RUN echo "cd ${CHLOE_DIR} && ./chloe" >> ${HOME}/start.sh
RUN chmod +x ${HOME}/start.sh

ENTRYPOINT [ "/bin/sh", "/root/start.sh" ]

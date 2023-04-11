FROM python:3.10.11-alpine

ARG DETECTLANG_APIKEY

ARG HOME=/root
ARG GOPATH=${HOME}/go
ARG CHLOE_DIR=${GOPATH}/src/chloe

ENV GOPATH=${GOPATH}

WORKDIR ${CHLOE_DIR}

RUN apk update
RUN apk add go git protoc ffmpeg

RUN mkdir -p ${HOME}/go/src

RUN git clone https://github.com/DiamondGo/Chloe.git ${CHLOE_DIR}
RUN cd ${CHLOE_DIR} && git submodule update --init --recursive

RUN cd ${CHLOE_DIR}/pyservice && pip install -r requirements.txt \
    && python -m grpc_tools.protoc -I./proto --python_out=pb --grpc_python_out=pb proto/tts.proto
RUN cd ${CHLOE_DIR} && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest \
    && export PATH=${GOPATH}/bin:$PATH \
    && protoc --go_out=. --go-grpc_out=. pyservice/proto/tts.proto
RUN cd ${CHLOE_DIR} && go build
RUN mkdir -p ${CHLOE_DIR}/log

RUN echo "cd ${CHLOE_DIR}/pyservice && python main.py ${DETECTLANG_APIKEY} pyservice &" >> ${HOME}/start.sh
RUN echo "cd ${CHLOE_DIR} && ./chloe" >> ${HOME}/start.sh
RUN chmod +x ${HOME}/start.sh

ENTRYPOINT [ "/bin/sh", "/root/start.sh" ]

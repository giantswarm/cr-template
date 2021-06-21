FROM quay.io/giantswarm/alpine:3.14 AS binaries

ARG KUBECTL_VERSION=1.29.2

RUN apk add --no-cache ca-certificates curl \
    && mkdir -p /binaries \
    && curl -SL https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl -o /binaries/kubectl \
    && chmod +x /binaries/*

FROM quay.io/giantswarm/alpine:3.14

COPY --from=binaries /binaries/* /usr/bin/
COPY ./kubectl-gs /usr/bin/kubectl-gs

ENTRYPOINT ["kubectl-gs"]

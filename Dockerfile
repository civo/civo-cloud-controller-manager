FROM golang:1.16.5 as builder

ARG ccm_version=dev

RUN mkdir /src
ADD . /src/
WORKDIR /src
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -ldflags "-s -w" -X github.com/civo/civo-cloud-controller-manager/cloud-controller-manager/civo.CCMVersion=${ccm_version} -o civo-cloud-controller-manager github.com/civo/civo-cloud-controller-manager/cloud-controller-manager/cmd/civo-cloud-controller-manager
RUN ls

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /src/civo-cloud-controller-manager /civo-cloud-controller-manager
ENTRYPOINT ["/civo-cloud-controller-manager"]


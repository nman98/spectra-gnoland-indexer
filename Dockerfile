FROM golang:1.25 AS builder

WORKDIR /app

COPY . .

ARG GIT_COMMIT=""
ARG GIT_TAG=""
ARG GIT_BRANCH=""
ARG VERSION=""

RUN if [ -z "$VERSION" ]; then \
    if [ -n "$GIT_TAG" ]; then VERSION="$GIT_TAG"; \
    else VERSION="${GIT_BRANCH}-${GIT_COMMIT}"; \
    fi; \
    fi && \
    go build -ldflags="-s -w \
        -X 'github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/cmd.Commit=${GIT_COMMIT}' \
        -X 'github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/cmd.Version=${VERSION}'" \
        -o indexer ./indexer

RUN chmod +x indexer
RUN touch config.yml

FROM gcr.io/distroless/base-debian13:latest

WORKDIR /app

COPY --from=builder /app/indexer .
COPY --from=builder --chown=nonroot:nonroot /app/config.yml .

USER nonroot

ENTRYPOINT ["/app/indexer"]
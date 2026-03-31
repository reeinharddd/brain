FROM qdrant/qdrant:v1.10.1

# Ensure we can run the healthcheck which uses curl
USER root
RUN set -eux; \
    if command -v apk >/dev/null 2>&1; then \
      apk add --no-cache curl; \
    elif command -v apt-get >/dev/null 2>&1; then \
      apt-get update && apt-get install -y --no-install-recommends curl ca-certificates && rm -rf /var/lib/apt/lists/*; \
    elif command -v yum >/dev/null 2>&1; then \
      yum install -y curl ca-certificates && yum clean all; \
    else \
      echo "No known package manager to install curl" >&2; exit 1; \
    fi

# Keep upstream ENTRYPOINT/CMD

FROM node:20-alpine

WORKDIR /app

# Install Docker CLI (needed to spawn Docker-based MCPs)
# Use package 'docker' (Alpine package name) instead of 'docker-cli' and
# add curl. This is more widely available across Alpine versions.
RUN apk add --no-cache docker curl

# Copy package files
COPY package.json ./
RUN npm install --production

# Copy server
COPY server.js ./

# Create data directory
RUN mkdir -p /data

ENV PORT=3000
ENV MCP_REGISTRY=/app/mcp-registry.json
ENV MCP_DATA_DIR=/data

EXPOSE 3000

CMD ["node", "server.js"]

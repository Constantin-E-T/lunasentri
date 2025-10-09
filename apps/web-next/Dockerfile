FROM node:20-alpine

WORKDIR /app

# Install system dependencies
RUN apk add --no-cache curl \
  && corepack enable \
  && corepack prepare pnpm@10.15.1 --activate

# Copy package files
COPY package*.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# Copy rest of the app
COPY . .

# Build arguments and environment
ARG NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
ENV NODE_ENV=production

# Build the app
RUN pnpm run build

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
  CMD curl -fsS http://localhost:80/api/health || exit 1

CMD ["pnpm", "run", "start:prod"]
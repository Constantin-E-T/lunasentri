FROM node:20-alpine

WORKDIR /app

# Install system dependencies and enable pnpm
RUN apk add --no-cache curl \
  && corepack enable \
  && corepack prepare pnpm@10.18.0 --activate

# Copy package files
COPY package*.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# Copy rest of the app
COPY . .

# Declare build arguments for environment variables
ARG NEXT_PUBLIC_API_URL

# Set environment variables from build arguments
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
ENV NODE_ENV=production

# Build the app with environment variables available
RUN pnpm run build

EXPOSE 80

# Health check for CapRover
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
  CMD curl -fsS http://localhost:80/ || exit 1

# Start Next.js on port 80
CMD ["pnpm", "run", "start", "-p", "80"]
FROM node:20-alpine

WORKDIR /app/ui

RUN npm install -g pnpm

# Copy package files first to leverage Docker caching
COPY package.json pnpm-lock.yaml ./

RUN pnpm install

# Copy the rest of the UI code
COPY . ./

# Build the UI
RUN pnpm run build

# Install serve to serve the static files
RUN npm install -g serve

# Add labels
LABEL org.opencontainers.image.title="Olake Frontend"
LABEL org.opencontainers.image.description="Frontend UI for Olake application"

# Expose port for the UI
EXPOSE 5173

# Command to serve the UI
CMD ["serve", "-s", "dist", "-l", "5173"] 
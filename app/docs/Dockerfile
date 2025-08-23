FROM node:18-alpine

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production

# Copy source code
COPY . .

# Build the documentation
RUN npm run build

# Expose port
EXPOSE 3000

# Start the documentation server
CMD ["npm", "run", "serve", "--", "--host", "0.0.0.0", "--port", "3000"]

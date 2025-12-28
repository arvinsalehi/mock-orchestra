FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
# Vite dev server runs on 5173 by default, or 3000 if configured
EXPOSE 3000 
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0", "--port", "3000"]

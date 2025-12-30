FROM node:18-alpine

WORKDIR /app

# install production deps first
COPY package.json package-lock.json* ./
RUN npm install --only=production

# copy source
COPY . .

ENV NODE_ENV=production
ENV PORT=3000

EXPOSE 3000

USER node

CMD ["node", "src/server.js"]

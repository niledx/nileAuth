FROM node:18-alpine

WORKDIR /app

# install production deps first
COPY node/package.json node/package-lock.json* ./
RUN npm install --only=production

# copy source
COPY node/ .

ENV NODE_ENV=production
ENV PORT=3000

EXPOSE 3000

USER node

CMD ["node", "src/server.js"]

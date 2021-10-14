FROM node:lts as builder

WORKDIR /usr/src/app

COPY . .

RUN npm install

RUN npm run build

# Without dev dependencies

FROM node:lts-alpine

WORKDIR /usr/src/app

ENV NODE_ENV=production

COPY package.json .

RUN npm install --only=production

COPY --from=builder /usr/src/app/dist .

CMD ["node", "--expose_gc", "index.js"]
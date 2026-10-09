# syntax=docker/dockerfile:1

FROM node:20-alpine AS dependencies
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

FROM node:20-alpine AS builder
WORKDIR /app
ARG SUB_APP_TYPE=admin
ARG ADMIN_API_BASE_URL=
ARG CUSTOMER_API_BASE_URL=
ARG NEXT_PUBLIC_API_BASE_URL=
ARG NEXT_PUBLIC_ACCOUNT_REGISTER_PATH=/auth/accounts/register
ARG NEXT_PUBLIC_LOGIN_PATH=/auth/login
ENV NEXT_TELEMETRY_DISABLED=1 \
    SUB_APP_TYPE=${SUB_APP_TYPE} \
    ADMIN_API_BASE_URL=${ADMIN_API_BASE_URL} \
    CUSTOMER_API_BASE_URL=${CUSTOMER_API_BASE_URL} \
    NEXT_PUBLIC_API_BASE_URL=${NEXT_PUBLIC_API_BASE_URL} \
    NEXT_PUBLIC_ACCOUNT_REGISTER_PATH=${NEXT_PUBLIC_ACCOUNT_REGISTER_PATH} \
    NEXT_PUBLIC_LOGIN_PATH=${NEXT_PUBLIC_LOGIN_PATH}
COPY --from=dependencies /app/node_modules ./node_modules
COPY . .
RUN mkdir -p public && npm run build

FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production \
    NEXT_TELEMETRY_DISABLED=1 \
    PORT=3000 \
    HOSTNAME="0.0.0.0"

RUN addgroup --system --gid 1001 nodejs \
    && adduser --system --uid 1001 nextjs

COPY --from=builder --chown=nextjs:nodejs /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs
EXPOSE 3000
CMD ["node", "server.js"]

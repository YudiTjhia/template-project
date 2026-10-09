# syntax=docker/dockerfile:1

FROM node:24-alpine AS dependencies
WORKDIR /app
COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci --fetch-retries=5 --maxsockets=1

FROM node:24-alpine AS builder
WORKDIR /app
ARG SUB_APP_TYPE=admin
ARG ADMIN_API_BASE_URL=
ARG CUSTOMER_API_BASE_URL=
ARG NEXT_PUBLIC_API_BASE_URL=
ARG NEXT_PUBLIC_ACCOUNT_REGISTER_PATH=/auth/accounts/register
ARG NEXT_PUBLIC_LOGIN_PATH=/auth/login
ARG NEXT_PUBLIC_FIREBASE_API_KEY=
ARG NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN=
ARG NEXT_PUBLIC_FIREBASE_PROJECT_ID=
ARG NEXT_PUBLIC_FIREBASE_STORAGE_BUCKET=
ARG NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID=
ARG NEXT_PUBLIC_FIREBASE_APP_ID=
ARG NEXT_PUBLIC_FIREBASE_VAPID_KEY=
ENV NEXT_TELEMETRY_DISABLED=1 \
    SUB_APP_TYPE=${SUB_APP_TYPE} \
    NEXT_PUBLIC_SUB_APP_TYPE=${SUB_APP_TYPE} \
    ADMIN_API_BASE_URL=${ADMIN_API_BASE_URL} \
    CUSTOMER_API_BASE_URL=${CUSTOMER_API_BASE_URL} \
    NEXT_PUBLIC_API_BASE_URL=${NEXT_PUBLIC_API_BASE_URL} \
    NEXT_PUBLIC_ACCOUNT_REGISTER_PATH=${NEXT_PUBLIC_ACCOUNT_REGISTER_PATH} \
    NEXT_PUBLIC_LOGIN_PATH=${NEXT_PUBLIC_LOGIN_PATH} \
    NEXT_PUBLIC_FIREBASE_API_KEY=${NEXT_PUBLIC_FIREBASE_API_KEY} \
    NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN=${NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN} \
    NEXT_PUBLIC_FIREBASE_PROJECT_ID=${NEXT_PUBLIC_FIREBASE_PROJECT_ID} \
    NEXT_PUBLIC_FIREBASE_STORAGE_BUCKET=${NEXT_PUBLIC_FIREBASE_STORAGE_BUCKET} \
    NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID=${NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID} \
    NEXT_PUBLIC_FIREBASE_APP_ID=${NEXT_PUBLIC_FIREBASE_APP_ID} \
    NEXT_PUBLIC_FIREBASE_VAPID_KEY=${NEXT_PUBLIC_FIREBASE_VAPID_KEY}
COPY --from=dependencies /app/node_modules ./node_modules
COPY . .
# Persist Next's compiler cache between builds without including it in the
# runtime image.
RUN --mount=type=cache,target=/app/.next/cache mkdir -p public \
    && npm run build

FROM node:24-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production \
    NEXT_TELEMETRY_DISABLED=1 \
    PORT=3000 \
    HOSTNAME="0.0.0.0"

RUN addgroup --system --gid 1001 nodejs \
    && adduser --system --uid 1001 nextjs \
    && mkdir -p .next \
    && chown nextjs:nodejs .next

COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static
# A public directory is optional in Next.js projects.
COPY --from=builder --chown=nextjs:nodejs /app/public ./public

USER nextjs
EXPOSE 3000
CMD ["node", "server.js"]

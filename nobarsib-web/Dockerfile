# syntax=docker/dockerfile:1

FROM node:20-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

FROM node:20-alpine AS build
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

# NEXT_PUBLIC_* ditanam saat build, bukan saat container jalan: nilainya ikut
# masuk ke bundel JavaScript yang diunduh browser. Alamat ini harus yang
# terlihat dari luar container, bukan nama service Docker.
ARG NEXT_PUBLIC_API_URL=http://localhost:8080/v1
ARG NEXT_PUBLIC_ADMIN_WA=
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL \
    NEXT_PUBLIC_ADMIN_WA=$NEXT_PUBLIC_ADMIN_WA \
    NEXT_TELEMETRY_DISABLED=1

# .next/cache adalah cache build Turbopack — ratusan megabita yang tidak
# berguna sama sekali saat runtime. npm prune membuang dependensi
# pengembangan yang juga tidak dipakai server.
RUN npm run build \
    && rm -rf .next/cache \
    && npm prune --omit=dev

FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production \
    NEXT_TELEMETRY_DISABLED=1 \
    TZ=Asia/Jakarta
RUN apk add --no-cache tzdata && adduser -D -u 10001 nextjs

# Dijalankan dengan `next start`, BUKAN keluaran `output: "standalone"`.
#
# Standalone menghasilkan image jauh lebih kecil, tapi server minimalnya tidak
# menerapkan proxy.ts maupun headers() dari next.config.ts — sudah diverifikasi
# pada build yang sama: `next start` mengirim header keamanan, `node server.js`
# dari keluaran standalone tidak mengirim apa pun.
#
# Gagalnya diam-diam, tanpa error, dan itulah yang membuatnya berbahaya:
# portal venue di Fase 5 akan memakai proxy untuk melindungi rute, dan
# perlindungan yang tidak jalan di produksi tanpa memberi tanda jauh lebih mahal
# daripada beberapa ratus megabita image.
COPY --from=build /app/node_modules ./node_modules
COPY --from=build /app/.next ./.next
COPY --from=build /app/public ./public
COPY --from=build /app/package.json ./package.json
COPY --from=build /app/next.config.ts ./next.config.ts

USER nextjs
EXPOSE 3000
ENV PORT=3000 HOSTNAME=0.0.0.0
CMD ["npm", "start"]

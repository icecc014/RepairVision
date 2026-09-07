FROM node:20.16.0-alpine AS pc-builder
WORKDIR /src/pc
COPY frontend/pc-admin/package.json frontend/pc-admin/package-lock.json ./
RUN npm ci
COPY frontend/pc-admin/ ./
RUN npm run build

FROM node:20.16.0-alpine AS mobile-builder
WORKDIR /src/m
COPY frontend/mobile-web/package.json frontend/mobile-web/package-lock.json ./
RUN npm ci
COPY frontend/mobile-web/ ./
RUN npm run build

FROM nginx:1.27.0-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=pc-builder /src/pc/dist /usr/share/nginx/html/admin
COPY --from=mobile-builder /src/m/dist /usr/share/nginx/html/m
EXPOSE 80
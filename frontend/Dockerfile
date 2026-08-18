FROM registry-1.docker.io/library/node:22-alpine AS admin-builder

WORKDIR /src/admin

COPY admin/package.json admin/package-lock.json ./
RUN npm ci

COPY admin/ ./
RUN npm run build

FROM registry-1.docker.io/library/nginx:1.28-alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY index/ /usr/share/nginx/html/index/
COPY downloads/ /usr/share/nginx/html/index/downloads/
COPY --from=admin-builder /src/admin/dist/ /usr/share/nginx/html/admin/

EXPOSE 80 8080

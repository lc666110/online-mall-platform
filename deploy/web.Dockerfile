FROM node:22-alpine AS build
WORKDIR /build
COPY package.json package-lock.json ./
COPY frontend/package.json frontend/package.json
COPY admin/package.json admin/package.json
RUN npm ci --no-audit --no-fund
COPY shared shared
COPY frontend frontend
COPY admin admin
ARG APP=frontend
RUN npm run build -w ${APP} && cp -r ${APP}/dist /site
FROM nginx:1.29-alpine
COPY deploy/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /site /usr/share/nginx/html
EXPOSE 80
HEALTHCHECK --interval=15s --timeout=3s CMD wget -qO- http://127.0.0.1/ >/dev/null || exit 1

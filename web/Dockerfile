# syntax=docker/dockerfile:1

# ---- build: compile the Vite SPA ----
FROM node:22 AS build
WORKDIR /app

# Install deps first (cached across source-only changes). npm ci needs the
# lockfile and installs devDependencies too, which the build (tsc + vite) needs.
COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run build

# ---- serve: nginx serving the static bundle + reverse-proxying the API ----
FROM nginx:alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80

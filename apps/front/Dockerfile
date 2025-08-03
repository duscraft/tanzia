FROM nginx:latest

WORKDIR /app

COPY dist .
COPY nginx.conf /etc/nginx/nginx.conf
USER nginx

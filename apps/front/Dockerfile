FROM nginx:latest

COPY dist/front/browser/ /usr/share/nginx/html
COPY apps/front/nginx.conf /etc/nginx/nginx.conf
USER nginx

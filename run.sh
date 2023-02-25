docker build -t lebot . --tag lebot:latest

docker run -d \
  --name lebot \
  --mount type=bind,source="/apps/lebot/logs",target=/tmp/logs/lebot \
  --mount type=bind,source="/apps/lebot/secrets",target=/tmp/secrets \
  --restart always \
  lebot:latest

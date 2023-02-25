docker build -t lebot .

docker run -d \
  --name lebot \
  --mount type=bind,source="/tmp/logs/lebot",target=/tmp/logs/lebot \
  --mount type=bind,source="/tmp/secrets",target=/tmp/secrets \
  --restart always
  lebot:latest

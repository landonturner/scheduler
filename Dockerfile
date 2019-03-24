FROM golang

WORKDIR src/bitbucket.org/maprtech/build-scheduler-api
COPY . .

CMD ["env", "JWT_SECRET=d7d3b8dfda6bbcb893f962d195ec387731f1483aad63ed283ad699cbc922b8c6", "go", "run", "server.go"]

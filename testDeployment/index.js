const http = require("http");

const server = http.createServer((req, res) => {
  res.end("ok");
});

server.listen(8080);

module.exports = server;

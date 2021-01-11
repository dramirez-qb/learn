// Import the HTTP module
const http = require("http");
// Import the URL module
const url = require("url");

const hostname = '0.0.0.0';
const port = 3000;

// Make our HTTP server
const server = http.createServer((req, res) => {
  // Set our header
  res.setHeader("Access-Control-Allow-Origin", "*")
  res.setHeader('Content-Type', 'text/plain');
  res.statusCode = 200;
  // Parse the request url
  const reqUrl = url.parse(req.url).pathname
  if (reqUrl == "/") {
    res.write('Hello World')
    res.end()
  } else if (reqUrl == "/ping") {
    res.write("pong")
    res.end()
  } else if (reqUrl == "/healthz") {
    res.statusCode = 200;
    res.setHeader('Content-Type', 'application/json');
    res.write("{\"alive\": true}")
    res.end()
  } else {
    res.statusCode = 404;
    res.write("The page you are trying to reach either does not exist or you are not authorized to view it.")
    res.end()
  }
})

server.listen(port, hostname, () => {
  console.log(`Server running at http://${hostname}:${port}/`);
});

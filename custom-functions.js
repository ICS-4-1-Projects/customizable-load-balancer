// Filename: custom-functions.js
module.exports = {
  randomizePort: function (userContext, events, done) {
    // Randomly select a port from 8081, 8082, or 8083
    const ports = [8081, 8082, 8083];
    const randomPort = ports[Math.floor(Math.random() * ports.length)];
    userContext.vars.port = randomPort;
    done();
  },
  setPort: function (requestParams, userContext, events, done) {
    // Completely redefine the URL with the random port and endpoint
    requestParams.url = `http://localhost:${userContext.vars.port}${userContext.vars.endpoint}`;
    done();
  },
  setUserContextHello: function (userContext, events, done) {
    // Set the endpoint for "/hello"
    userContext.vars.endpoint = "/hello";
    done();
  },
  setUserContextHeartbeat: function (userContext, events, done) {
    // Set the endpoint for "/heartbeat"
    userContext.vars.endpoint = "/heartbeat";
    done();
  },
};

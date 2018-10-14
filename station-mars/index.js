'use strict';

// startTrace starts the trace
function startTrace(req) {
}

exports.http = (request, response) => {
  startTrace(request);

  let output = {
    planet: "mars",
    weather: "it's kinda cold here"
  };

  response.status(200).send(JSON.stringify(output));
};

exports.event = (event, callback) => {
  callback();
};

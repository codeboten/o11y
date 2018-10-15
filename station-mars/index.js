'use strict';

var beeline = require("honeycomb-beeline")({
  writeKey: process.env.HONEYCOMB_KEY,
  dataset: process.env.HONEYCOMB_DATASET
});


// startTrace returns a trace object
function startTrace(req) {
  let traceInfo = {
    application: "station-api",
    platform: "gcp",
    name: "handleRequest"
  };

  if (req.headers["x-honeycomb-trace"]) {
    let { traceId, parentSpanId } = beeline.unmarshalTraceContext(
      req.headers["x-honeycomb-trace"]
    );
    return beeline.startTrace(
      traceInfo,
      traceId,
      parentSpanId
    );
  }
  return beeline.startTrace(traceInfo);
}

exports.http = (request, response) => {
  let trace = startTrace(request);

  let output = {
    planet: "mars",
    weather: "it's kinda cold here"
  };

  beeline.finishTrace(trace);
  response.status(200).send(JSON.stringify(output));
};

exports.event = (event, callback) => {
  callback();
};

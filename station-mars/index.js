'use strict';

var beeline = require("honeycomb-beeline")({
    writeKey: process.env.HONEYCOMB_KEY,
    dataset: process.env.HONEYCOMB_DATASET
});

var application = "station-api";
var platform = "gcp";

function getRandomInt(max) {
    return Math.floor(Math.random() * Math.floor(max));
}

function contactWeatherStation(planet) {
    var span = beeline.startSpan({
        application: application,
        platform: platform,
        name: "contactWeatherStation",
        planet: planet
    });
    var weatherReports = [
        "it's kinda cold here",
        "nothing but red skies today",
        "better stay inside the biodome today"
    ];
    beeline.finishSpan(span);
    return weatherReports[getRandomInt(weatherReports.length)];
}

function getWeather(planet) {
    var span = beeline.startSpan({
        application: application,
        platform: platform,
        name: "getWeather",
        planet: planet
    });
    var weather = contactWeatherStation(planet);
    beeline.finishSpan(span);
    return weather;
}

// startTrace returns a trace object
function startTrace(req) {
    var traceInfo = {
        application: application,
        platform: platform,
        name: "handleRequest"
    };

    if (req.headers["x-honeycomb-trace"]) {
        let {traceId, parentSpanId} = beeline.unmarshalTraceContext(
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
    let planet = process.env.PLANET
    let output = {
        planet: planet,
        weather: getWeather(planet)
    };

    beeline.finishTrace(trace);
    response.status(200).send(JSON.stringify(output));
};

exports.event = (event, callback) => {
    callback();
};

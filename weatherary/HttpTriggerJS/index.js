var beeline = require("honeycomb-beeline")({
    writeKey: process.env.HONEYCOMB_KEY,
    dataset: process.env.HONEYCOMB_DATASET,
});

function getDistance(planet) {
    let span = beeline.startSpan({
        task: "getDistance",
        planet
      });
    beeline.finishSpan(span)
    return 10000
}

function getWeather(planet) {
    let span = beeline.startSpan({
        task: "getWeather",
        planet
      });
    let distance = getDistance(planet)
    beeline.finishSpan(span)
    return "fine"
}

module.exports = async function (context, req) {
    let trace = beeline.startTrace({
        task: "httpTrigger"
    });
    context.log('JavaScript HTTP trigger function processed a request.');

    if (req.body && req.body.planet) {
        beeline.customContext.add("planet", );
        response = {
            "planet": req.body.planet,
            "weather": getWeather(req.body.planet),
        }
        context.res = {
            // status: 200, /* Defaults to 200 */
            body: JSON.stringify(response)
        };
    }
    else {
        beeline.
        context.res = {
            status: 400,
            body: "Please pass a planet in the request body"
        };
    }
    beeline.finishTrace(trace);
};
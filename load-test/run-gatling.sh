#!/usr/bin/env bash

RESULTS_WORKSPACE="$(pwd)/user-files/results"
GATLING_WORKSPACE="$(pwd)/user-files"

# check if directory gatling-charts-highcharts-bundle exists. if it doesnt, then download and unzip
if [ ! -d "gatling-charts-highcharts-bundle-3.9.5" ]; then
    echo "gatling-charts-highcharts-bundle-3.9.5 not found. Downloading..."
    wget --no-verbose https://repo1.maven.org/maven2/io/gatling/highcharts/gatling-charts-highcharts-bundle/3.9.5/gatling-charts-highcharts-bundle-3.9.5-bundle.zip
    unzip gatling-charts-highcharts-bundle-3.9.5-bundle.zip
fi

runGatling() {
    sh $cd gatling-charts-highcharts-bundle-3.9.5/bin/gatling.sh -rm local -s RinhaBackendCrebitosSimulation \
        -rd "Rinha de Backend - 2024/Q1: Crébito" \
        -rf $RESULTS_WORKSPACE \
        -sf "$GATLING_WORKSPACE/simulations"
}

startTest() {
    for i in {1..20}; do
        # 2 requests to wake the 2 api instances up :)
        curl --fail http://localhost:9999/clientes/1/extrato && \
        echo "" && \
        curl --fail http://localhost:9999/clientes/1/extrato && \
        echo "" && \
        runGatling && \
        break || sleep 2;
    done
}

startTest
#!/usr/bin/env bash

set -eu

if ! [ -d output ]; then
    mkdir output
fi

exit_code=0

echo 'Converting 3.1 spec to 3.0'
if docker run --rm -i openapi-spec-converter:latest -t 3.0 -f yaml \
    < specs/31-spec-with-differences-from-30.yaml \
    > output/31-spec-with-differences-from-30.converted-30.yaml
then
    echo 'Validating 3.1 spec converted to 3.0'
    if ! node_modules/.bin/swagger-cli validate output/31-spec-with-differences-from-30.converted-30.yaml; then
        exit_code=1
    fi

    echo 'Converting 3.1 spec to Swagger'
    if docker run --rm -i openapi-spec-converter:latest -t swagger -f yaml \
        < specs/31-spec-with-differences-from-30.yaml \
        > output/31-spec-with-differences-from-30.converted-swagger.yaml
    then
        echo 'Validating 3.1 spec converted to Swagger'
        if ! node_modules/.bin/swagger-cli validate output/31-spec-with-differences-from-30.converted-swagger.yaml; then
            exit_code=1
        fi

        # Up convert Swagger file back to OpenAPI 3.1 again, and output as JSON
        echo 'Converting 3.1 to Swagger spec back to 3.1 again'
        if docker run --rm -i openapi-spec-converter:latest -t 3.1 -f json \
            < output/31-spec-with-differences-from-30.converted-swagger.yaml \
            > output/31-spec-with-differences-from-30.back-to-31.json
        then
            echo 'Validating 3.1 spec converted back from Swagger'
            if ! node_modules/.bin/redocly lint output/31-spec-with-differences-from-30.back-to-31.json 2>&1; then
                exit_code=1
            fi
        else
            exit_code=1
        fi
    else
        exit_code=1
    fi
else
    exit_code=1
fi

echo 'Converting OpenAI OpenAPI 3.0 spec to 3.1'
if docker run --rm -i openapi-spec-converter:latest -t 3.1 -f json \
    < specs/openapi.yaml \
    > output/openapi-31.json
then
    echo 'Validating OpenAI OpenAPI spec converted to 3.0'
    if ! node_modules/.bin/swagger-cli validate output/openapi-31.json; then
        exit_code=1
    fi
else
    exit_code=1
fi

echo 'Converting OpenAI OpenAPI 3.0 spec to Swagger'
if docker run --rm -i openapi-spec-converter:latest -t swagger -f json \
    < specs/openapi.yaml \
    > output/openapi-swagger.json
then
    echo 'Validating OpenAI OpenAPI spec converted to Swagger'
    if ! node_modules/.bin/swagger-cli validate output/openapi-swagger.json; then
        exit_code=1
    fi
else
    exit_code=1
fi

exit $exit_code

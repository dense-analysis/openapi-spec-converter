#!/usr/bin/env bash

set -eu

if ! [ -d output ]; then
    mkdir output
fi

# Read in stdin and output to stdout
echo 'Converting 3.1 spec to 3.0'
docker run --rm -i openapi-spec-converter:latest -t 3.0 -f yaml \
    < specs/31-spec-with-differences-from-30.yaml \
    > output/31-spec-with-differences-from-30.converted-30.yaml

exit_code=0

echo 'Validating 3.1 spec converted to 3.0'
if ! node_modules/.bin/swagger-cli validate output/31-spec-with-differences-from-30.converted-30.yaml; then
    exit_code=1
fi

echo 'Converting 3.1 spec to Swagger'
# Read in from a filename and output to a filename
docker run --volume ./specs:/specs:ro --volume ./output:/output --rm \
    -i openapi-spec-converter:latest -t swagger -f yaml \
    -o /output/31-spec-with-differences-from-30.converted-swagger.yaml \
    /specs/31-spec-with-differences-from-30.yaml

echo 'Validating 3.1 spec converted to Swagger'
if ! node_modules/.bin/swagger-cli validate output/31-spec-with-differences-from-30.converted-swagger.yaml; then
    exit_code=1
fi

echo 'Converting 3.1 to Swagger spec back to 3.1 again'
# Up convert Swagger file back to OpenAPI 3.1 again, and output as JSON
docker run --volume ./output:/output \
    --rm -i openapi-spec-converter:latest -t 3.1 -f json \
    -o /output/31-spec-with-differences-from-30.back-to-31.yaml \
    /output/31-spec-with-differences-from-30.converted-swagger.yaml

echo 'Validating 3.1 spec converted back from Swagger'
if ! node_modules/.bin/redocly lint output/31-spec-with-differences-from-30.converted-swagger.yaml 2>&1; then
    exit_code=1
fi

exit $exit_code

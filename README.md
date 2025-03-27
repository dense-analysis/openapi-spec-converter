# openapi-spec-converter

openapi-spec-converter converts between Swagger, OpenAPI 3.0, and OpenAPI 3.1
documents in JSON or YAML formats. This tool exists to bridge gaps between
tools for various languages for generating client code where support for any
of the above document types is inconsistent.

## Usage

Run the spec converting with Docker like so:

```sh
docker run --rm -i openapi-spec-converter:latest
```

At the time of writing the following options are supported.

```text
Usage: openapi-spec-converter [-h] [-f value] [-o value] [-t value] <input>
 -f, --format=value
             Output format: yaml or json [json]
 -h, --help  Print this help message
 -o, --output=value
             Output file (default stdout)
 -t, --target=value
             Target version: swagger, 3.0, or 3.1 [3.1]
```

The input file can be specified as `-` for stdin, or omitted if piping in a
file. In the simplest usage, you might want to do the following to get a valid
OpenAPI 3.1 spec from any format.

```sh
docker run --rm -i openapi-spec-converter:latest < file.json
```

The spec converter will output to JSON by default. You can pass `-f yaml` to
change the output format to YAML.

## Development

You can build the Docker image with the following command.

```sh
docker build -t openapi-spec-converter:latest .
```

If you want to work on the code locally, insure you have Go 1.24 installed,
and use `gopls` as your language server.

```sh
go install "golang.org/dl/go1.24.1@latest"
go1.24.1 download
go install golang.org/x/tools/gopls@latest
```

You can run the program directly while testing with `go run`.

To run tests for the program, you'll need a recent Node and `npm` version. You
can use the latest version with `nvm` like so.

```sh
nvm install node
nvm use node
```

```sh
# Ensure you have a somewhat recent version of Node and npm installed first.
npm install
./convert-and-validate-specs
```

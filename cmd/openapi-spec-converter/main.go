package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	openapispecconverter "github.com/dense-analysis/openapi-spec-converter"
	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	ghodssYaml "github.com/ghodss/yaml"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/datamodel/low"
	"github.com/pb33f/libopenapi/utils"
	"github.com/pborman/getopt/v2"
	"gopkg.in/yaml.v3"
)

type SpecVersion int

const (
	Swagger SpecVersion = iota
	OpenAPI30
	OpenAPI31
)

type Format int

const (
	JSON Format = iota
	YAML
)

type Arguments struct {
	inputFilename  string
	outputFilename string
	outputTarget   SpecVersion
	outputFormat   Format
}

func parseArgs() Arguments {
	var arguments Arguments

	getopt.SetProgram(filepath.Base(os.Args[0]))

	showHelp := getopt.BoolLong("help", 'h', "Print this help message")
	outputFilename := getopt.StringLong("output", 'o', "", "Output file (default stdout)")
	outputVersion := getopt.StringLong("target", 't', "3.1", "Target version: swagger, 3.0, or 3.1")
	outputFormat := getopt.StringLong("format", 'f', "json", "Output format: yaml or json")
	getopt.SetParameters("<input>")

	getopt.Parse()

	if showHelp != nil && *showHelp {
		getopt.PrintUsage(os.Stdout)
		os.Exit(0)
	}

	args := getopt.Args()

	if len(args) > 2 {
		fmt.Fprintln(os.Stderr, "Invalid number of arguments")
		getopt.PrintUsage(os.Stderr)
		os.Exit(1)
	}

	if len(args) == 0 {
		// If no arguments are supplied and there's no data being piped in,
		// then complain and print usage.
		if stat, err := os.Stdin.Stat(); err != nil || (stat.Mode()&os.ModeCharDevice) != 0 {
			fmt.Fprintln(os.Stderr, "No input filename or open stdin pipe")
			getopt.PrintUsage(os.Stderr)
			os.Exit(1)
		}

		arguments.inputFilename = "-"
	} else {
		arguments.inputFilename = args[0]
	}

	if len(arguments.inputFilename) == 0 {
		fmt.Fprintln(os.Stderr, "Empty input filename")
		getopt.PrintUsage(os.Stderr)
		os.Exit(1)
	}

	arguments.outputFilename = *outputFilename

	switch strings.ToLower(*outputVersion) {
	case "swagger":
		arguments.outputTarget = Swagger
	case "3.0":
		arguments.outputTarget = OpenAPI30
	case "3.1":
		arguments.outputTarget = OpenAPI31
	default:
		fmt.Fprintf(os.Stderr, "Invalid target version %s\n", *outputVersion)
		getopt.PrintUsage(os.Stderr)
		os.Exit(1)
	}

	switch strings.ToLower(*outputFormat) {
	case "json":
		arguments.outputFormat = JSON
	case "yaml":
		arguments.outputFormat = YAML
	default:
		fmt.Fprintf(os.Stderr, "Invalid format: %s\n", *outputFormat)
		getopt.PrintUsage(os.Stderr)
		os.Exit(1)
	}

	return arguments
}

func readInputFile(arguments Arguments) (inputData []byte, err error) {
	if arguments.inputFilename == "-" {
		inputData, err = io.ReadAll(os.Stdin)
	} else {
		inputData, err = os.ReadFile(arguments.inputFilename)
	}

	return
}

func make30RequiredAndReadonlyPropertiesOnlyReadonly(schema *base.Schema) {
	if schema.Properties != nil && len(schema.Required) > 0 {
		newRequired := []string{}

		for _, propName := range schema.Required {
			readonly := false

			if schema.Properties != nil {
				if item, ok := schema.Properties.Get(propName); ok {
					propSchema := item.Schema()

					readonly = propSchema.ReadOnly != nil && *propSchema.ReadOnly
				}
			}

			if !readonly {
				newRequired = append(newRequired, propName)
			}
		}

		schema.Required = newRequired
	}
}

func convert30NullablesTo31TypeArrays(schema *base.Schema) {
	// Replace {type: T, nullable: true} with {type: [T, "null"]}, etc.
	if schema.Nullable != nil {
		if *schema.Nullable {
			schema.Type = append(schema.Type, "null")
		}

		schema.Nullable = nil
	}
}

func convert31TypeArraysTo30(schema *base.Schema) {
	nullable := false
	nonNullType := ""

	for _, value := range schema.Type {
		if value == "null" {
			nullable = true
		} else {
			nonNullType = value
		}
	}

	if nullable && len(schema.Type) == 2 {
		// In case of {type: [T, "null"]} set {type: T, nullable: true}
		schema.Type[0] = nonNullType
		schema.Type = schema.Type[:1]
		schema.Nullable = &nullable
	} else if len(schema.Type) >= 2 {
		// In case of 2 or more non-null values, set them in oneOf
		// if "null" was one of the values then all values will be nullable.
		schema.OneOf = make([]*base.SchemaProxy, 0, len(schema.Type))

		for _, value := range schema.Type {
			if value != "null" {
				newSchema := base.Schema{Type: []string{value}}

				if nullable {
					newSchema.Nullable = &nullable
				}

				schema.OneOf = append(schema.OneOf, base.CreateSchemaProxy(&newSchema))
			}
		}

		// Clear the type field.
		schema.Type = nil
	}
}

func convert30MinMaxTo31(schema *base.Schema) {
	convert30ExclusiveBoundTo31 := func(
		bound **float64,
		exclusiveBound **base.DynamicValue[bool, float64],
	) {
		if *exclusiveBound != nil && (*exclusiveBound).IsA() {
			if (*exclusiveBound).A {
				// Before: {miniumum: val, exclusiveMinimum: true}
				// After: {exclusiveMinimum: val}
				if *bound != nil {
					(*exclusiveBound).N = 1
					(*exclusiveBound).B = **bound
				}

				*bound = nil
			} else {
				// Before: {minimum: val, exclusiveMinimum: false}
				// After: {minimum: val}
				*exclusiveBound = nil
			}
		}
	}

	convert30ExclusiveBoundTo31(&schema.Minimum, &schema.ExclusiveMinimum)
	convert30ExclusiveBoundTo31(&schema.Maximum, &schema.ExclusiveMaximum)
}

func convert31MinMaxTo30(schema *base.Schema) {
	convert31ExclusiveBoundTo30 := func(
		bound **float64,
		exclusiveBound **base.DynamicValue[bool, float64],
	) {
		if *exclusiveBound != nil && (*exclusiveBound).IsB() {
			// Before: {exclusiveMinimum: val}
			// After: {minimum: value, exclusiveMinimum: true}
			*bound = &(*exclusiveBound).B
			(*exclusiveBound).A = true
			(*exclusiveBound).N = 0
		}
	}

	convert31ExclusiveBoundTo30(&schema.Minimum, &schema.ExclusiveMinimum)
	convert31ExclusiveBoundTo30(&schema.Maximum, &schema.ExclusiveMaximum)
}

func convert30ExampleTo31Examples(schema *base.Schema) {
	if schema.Example != nil {
		schema.Examples = []*yaml.Node{schema.Example}
		schema.Example = nil
	}
}

func convert31ExamplesTo30Example(schema *base.Schema) {
	if len(schema.Examples) >= 1 {
		schema.Example = schema.Examples[0]
		schema.Examples = nil
	}
}

func convert30FormatsTo31ContentFields(schema *base.Schema) {
	if len(schema.Type) == 1 && schema.Type[0] == "string" && len(schema.Format) > 0 {
		if schema.Format == "binary" || schema.Format == "byte" {
			lowSchema := schema.GoLow()

			if lowSchema != nil {
				lowSchema.ContentMediaType = low.NodeReference[string]{
					Value:     "base64",
					ValueNode: utils.CreateStringNode("base64"),
				}
			}
		} else if schema.Format == "base64" {
			lowSchema := schema.GoLow()

			if lowSchema != nil {
				lowSchema.ContentEncoding = low.NodeReference[string]{
					Value:     "base64",
					ValueNode: utils.CreateStringNode("base64"),
				}
			}
		}

		schema.Format = ""
	}
}

func convert31ContentFieldsTo30Formats(schema *base.Schema) {
	if len(schema.Type) == 1 && schema.Type[0] == "string" {
		lowSchema := schema.GoLow()

		if lowSchema != nil {
			if len(lowSchema.ContentMediaType.Value) > 0 {
				if lowSchema.ContentMediaType.Value == "application/octet-stream" {
					schema.Format = "binary"
				}

				lowSchema.ContentMediaType.Mutate("")
			}

			if len(lowSchema.ContentEncoding.Value) > 0 {
				if lowSchema.ContentEncoding.Value == "base64" {
					schema.Format = "base64"
				}

				lowSchema.ContentEncoding.Mutate("")
			}
		}
	}
}

func updateSchemaAndReferencedSchema(
	schema *base.Schema,
	callback func(schema *base.Schema),
) {
	if schema == nil {
		// Skip editing nil schema.
		return
	}

	// Handle schemas in properties.
	if schema.Properties != nil {
		for property := range schema.Properties.ValuesFromOldest() {
			callback(property.Schema())
		}
	}

	// Handle items if the schema is an array.
	if schema.Items != nil {
		if schema.Items.IsA() {
			callback(schema.Items.A.Schema())
		}
	}

	// Process composite schemas: allOf, oneOf, and anyOf.
	for _, subSchema := range schema.AllOf {
		callback(subSchema.Schema())
	}

	for _, subSchema := range schema.OneOf {
		callback(subSchema.Schema())
	}

	for _, subSchema := range schema.AnyOf {
		callback(subSchema.Schema())
	}

	// Modify this schema last, so our changes to schema are final.
	callback(schema)
}

// updateAllSchema Finds schema anywhere they are used in spec and updates them using the `callback`
func updateAllSchema(
	model *libopenapi.DocumentModel[v3.Document],
	callback func(schema *base.Schema),
) {
	if model.Model.Components != nil && model.Model.Components.Schemas != nil {
		for value := range model.Model.Components.Schemas.ValuesFromOldest() {
			updateSchemaAndReferencedSchema(value.Schema(), callback)
		}
	}

	if model.Model.Components != nil && model.Model.Components.Parameters != nil {
		for value := range model.Model.Components.Parameters.ValuesFromOldest() {
			updateSchemaAndReferencedSchema(value.Schema.Schema(), callback)
		}
	}

	if model.Model.Paths != nil && model.Model.Paths.PathItems != nil {
		for pathItem := range model.Model.Paths.PathItems.ValuesFromOldest() {
			for operation := range pathItem.GetOperations().ValuesFromOldest() {
				if operation.RequestBody != nil && operation.RequestBody.Content != nil {
					for content := range operation.RequestBody.Content.ValuesFromOldest() {
						updateSchemaAndReferencedSchema(content.Schema.Schema(), callback)
					}
				}

				if operation.Responses != nil && operation.Responses.Codes != nil {
					for code := range operation.Responses.Codes.ValuesFromOldest() {
						if code.Content != nil {
							for mediaType := range code.Content.ValuesFromOldest() {
								updateSchemaAndReferencedSchema(mediaType.Schema.Schema(), callback)
							}
						}
					}
				}
			}
		}
	}
}

func clear30RequestFileContentSchemaFor31(
	model *libopenapi.DocumentModel[v3.Document],
) {
	if model.Model.Paths != nil && model.Model.Paths.PathItems != nil {
		for pathItem := range model.Model.Paths.PathItems.ValuesFromOldest() {
			for operation := range pathItem.GetOperations().ValuesFromOldest() {
				if operation.RequestBody != nil && operation.RequestBody.Content != nil {
					// Clear the schema for application/octet-stream, as the type is implied.
					if content, ok := operation.RequestBody.Content.Get("application/octet-stream"); ok {
						content.Schema = nil
					}
				}
			}
		}
	}
}

func set31RequestFileContentSchemaFor30(
	model *libopenapi.DocumentModel[v3.Document],
) {
	if model.Model.Paths != nil && model.Model.Paths.PathItems != nil {
		for pathItem := range model.Model.Paths.PathItems.ValuesFromOldest() {
			for operation := range pathItem.GetOperations().ValuesFromOldest() {
				if operation.RequestBody != nil && operation.RequestBody.Content != nil {
					// Clear the schema for application/octet-stream, as the type is implied.
					if content, ok := operation.RequestBody.Content.Get("application/octet-stream"); ok {
						content.Schema = base.CreateSchemaProxy(&base.Schema{
							Type:   []string{"string"},
							Format: "binary",
						})
					}
				}
			}
		}
	}
}

func fixSwaggerOperationUploadFormat(operation *openapi2.Operation) {
	if operation != nil && slices.Contains(operation.Consumes, "application/octet-stream") {
		for _, param := range operation.Parameters {
			if param.In == "body" && param.Schema == nil {
				param.Schema = &openapi2.SchemaRef{
					Value: &openapi2.Schema{
						Type:   &openapi3.Types{"string"},
						Format: "binary",
					},
				}
			}
		}
	}
}

func fixSwaggerDocUploadFormats(kinSwaggerDoc *openapi2.T) {
	for _, path := range kinSwaggerDoc.Paths {
		// HEAD, GET, DELETE we don't check here.
		// All other operations we try to fix.
		fixSwaggerOperationUploadFormat(path.Post)
		fixSwaggerOperationUploadFormat(path.Options)
		fixSwaggerOperationUploadFormat(path.Patch)
		fixSwaggerOperationUploadFormat(path.Put)
	}
}

func convertSwaggerToOpenAPI30(data []byte) ([]byte, error) {
	var kinSwaggerDoc openapi2.T

	dataFormat := checkDataFormat(data)

	// kin-openapi cannot unmarshal YAML correctly, so we have to first convert input to JSON.
	if dataFormat != JSON {
		var err error
		data, err = ghodssYaml.YAMLToJSON(data)

		if err != nil {
			return nil, fmt.Errorf("Error converting Swagger YAML to JSON: %w", err)
		}
	}

	if err := openapispecconverter.UnmarshalSwagger(data, &kinSwaggerDoc); err != nil {
		return nil, fmt.Errorf("Error loading Swagger data: %w", err)
	}

	if kinOpenAPIDoc, err := openapi2conv.ToV3(&kinSwaggerDoc); err == nil {
		return kinOpenAPIDoc.MarshalJSON()
	} else {
		return nil, fmt.Errorf("Error converting Swagger to 3.0 %w", err)
	}
}

func convertOpenAPI30ToSwagger(data []byte) ([]byte, error) {
	doc, err := libopenapi.NewDocument(data)

	if err != nil {
		return nil, fmt.Errorf("Error loading document: %w", err)
	}

	// Build the document in libopenapi so we can modify the document
	// to correct issues not handled by kin-openapi.
	model, errs := doc.BuildV3Model()

	if len(errs) > 0 {
		return nil, fmt.Errorf("Errors loading document: %w", errors.Join(errs...))
	}

	updateAllSchema(model, func(schema *base.Schema) {
		// We must make every property that is both required and also readonly
		// only be readonly, or they will break Swagger validation.
		make30RequiredAndReadonlyPropertiesOnlyReadonly(schema)
	})

	data, doc, model, errs = doc.RenderAndReload()

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	var kinSwaggerDoc *openapi2.T

	if kinOpenAPIDoc, err := openapi3.NewLoader().LoadFromData(data); err == nil {
		kinSwaggerDoc, err = openapispecconverter.OpenAPI30ToSwagger(kinOpenAPIDoc)

		if err != nil {
			return nil, fmt.Errorf("Error converting 3.0 to Swagger %w", err)
		}
	} else {
		return nil, fmt.Errorf("Error Load 3.0 for converting to Swagger %w", err)
	}

	// The kin-openapi Swagger converter doesn't add {schema: {type: "string", format: "binary"}}
	// when creating upload specs for binary content. We need to add it back in again.
	fixSwaggerDocUploadFormats(kinSwaggerDoc)

	return kinSwaggerDoc.MarshalJSON()
}

func convertOpenAPI30To31(data []byte) ([]byte, error) {
	doc, err := libopenapi.NewDocument(data)

	if err != nil {
		return nil, fmt.Errorf("Error loading document: %w", err)
	}

	model, errs := doc.BuildV3Model()

	if len(errs) > 0 {
		return nil, fmt.Errorf("Errors loading document: %w", errors.Join(errs...))
	}

	// See: https://www.openapis.org/blog/2021/02/16/migrating-from-openapi-3-0-to-3-1-0
	//
	// The following changes need to be made.
	//
	// 1. Change the `openapi` version to 3.1.x.
	// 2. Swap nullable for type arrays.
	// 3. Replace `minimum` and `exclusiveMinimum`, and `maximum` and `exclusiveMaximum`.
	// 4. Replace `example` with `examples` wherever we see it.
	// 5. Modify file upload schemas.

	// 1. Change the `openapi` version to 3.1.x.
	model.Model.Version = "3.1.1"

	// Before scanning all schema, apply step 5. early to clear schema for request bodies.
	clear30RequestFileContentSchemaFor31(model)

	updateAllSchema(model, func(schema *base.Schema) {
		// 2. Swap nullable for type arrays.
		convert30NullablesTo31TypeArrays(schema)
		// 3. Replace `minimum` and `exclusiveMinimum`
		convert30MinMaxTo31(schema)
		// 4. Replace `example` with `examples` wherever we see it.
		convert30ExampleTo31Examples(schema)
		// 5. Modify file upload schemas.
		convert30FormatsTo31ContentFields(schema)
	})

	data, doc, model, errs = doc.RenderAndReload()

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return data, nil
}

func convertOpenAPI31To30(data []byte) ([]byte, error) {
	doc, err := libopenapi.NewDocument(data)

	if err != nil {
		return nil, fmt.Errorf("Error loading document: %w", err)
	}

	model, errs := doc.BuildV3Model()

	if len(errs) > 0 {
		return nil, fmt.Errorf("Errors loading document: %w", errors.Join(errs...))
	}

	// We need to perform the inverse of the conversion steps in the 3.0 to 3.1 function.

	// 1. Change the `openapi` version to 3.0.x
	model.Model.Version = "3.0.4"

	// Before scanning all schema, apply step 5. early to schema schema for file uploads where needed.
	set31RequestFileContentSchemaFor30(model)

	updateAllSchema(model, func(schema *base.Schema) {
		// 2. Swap type arrays for either `nullable` or `oneOf`
		convert31TypeArraysTo30(schema)
		// 3. Replace `minimum` and `exclusiveMinimum`, and `maximum` and `exclusiveMaximum`.
		convert31MinMaxTo30(schema)
		// 4. Replace `examples` with `example` wherever we see it.
		convert31ExamplesTo30Example(schema)
		// 5. Modify file upload schemas.
		convert31ContentFieldsTo30Formats(schema)
	})

	// We must remove additional properties only used in 3.1.
	model.Model.JsonSchemaDialect = ""
	model.Model.Webhooks = nil

	if model.Model.Info != nil {
		model.Model.Info.Summary = ""
	}

	data, doc, model, errs = doc.RenderAndReload()

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return data, nil
}

func convertDocument(data []byte, outputVersion SpecVersion) ([]byte, error) {
	// First we'll parse the document in the simplest way to determine the document version.
	type BasicDoc struct {
		OpenAPI string `json:"openapi" yaml:"openapi"`
		Swagger string `json:"swagger" yaml:"swagger"`
	}
	var basicDoc BasicDoc

	if err := yaml.Unmarshal(data, &basicDoc); err != nil {
		return nil, fmt.Errorf("Cannot parse Swagger or OpenAPI document")
	}

	// Get the version string from the Swagger doc if empty.
	if len(basicDoc.OpenAPI) == 0 {
		basicDoc.OpenAPI = basicDoc.Swagger
	}

	// Build the model using libopenapi and determine the input version.
	var inputVersion SpecVersion

	switch basicDoc.OpenAPI {
	case "2.0":
		inputVersion = Swagger
	case "3.0.0", "3.0.1", "3.0.2", "3.0.3", "3.0.4":
		inputVersion = OpenAPI30
	case "3.1.0", "3.1.1":
		inputVersion = OpenAPI31
	default:
		return nil, fmt.Errorf("Unsuppoted input document OpenAPI version: %s", basicDoc.OpenAPI)
	}

	var err error

	// Cycle through document versions until we hit the one we want.
	for inputVersion != outputVersion {
		if inputVersion < outputVersion {
			if inputVersion == Swagger {
				data, err = convertSwaggerToOpenAPI30(data)
				inputVersion = OpenAPI30
			} else {
				data, err = convertOpenAPI30To31(data)
				inputVersion = OpenAPI31
			}
		} else {
			if inputVersion == OpenAPI31 {
				data, err = convertOpenAPI31To30(data)
				inputVersion = OpenAPI30
			} else {
				data, err = convertOpenAPI30ToSwagger(data)
				inputVersion = Swagger
			}
		}

		if err != nil {
			return nil, err
		}
	}

	return data, err
}

func checkDataFormat(data []byte) Format {
	for _, b := range data {
		switch b {
		case '{':
			return JSON
		case ' ', '\t', '\r', '\n':
		default:
			return YAML
		}
	}

	return YAML
}

func main() {
	arguments := parseArgs()

	data, err := readInputFile(arguments)

	if err != nil {
		log.Fatalf("Error reading input file %v\n", err)
	}

	data, err = convertDocument(data, arguments.outputTarget)

	if err != nil {
		log.Fatalf("Error converting document: %+v\n", err)
	}

	dataFormat := checkDataFormat(data)

	if dataFormat != arguments.outputFormat {
		if arguments.outputFormat == JSON {
			data, err = ghodssYaml.YAMLToJSON(data)
		} else {
			data, err = ghodssYaml.JSONToYAML(data)
		}

		if err != nil {
			log.Fatalf("Error converting to output format: %v\n", err)
		}
	}

	if len(arguments.outputFilename) > 0 {
		if err = os.WriteFile(arguments.outputFilename, data, 0644); err != nil {
			log.Fatalf("Error writing output file: %v\n", err)
		}
	} else {
		fmt.Println(string(data))
	}
}

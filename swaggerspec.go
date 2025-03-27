package openapispecconverter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
)

// This file redefines a Swagger spec and functions for unmarshalling and mapping to kin-openapi's structures for
// Swagger, simply because the spec that ships with the library isn't able to load its own output correctly,
// specfically when loading `type` keys.

// A substational amount of code from kin-openapi has been copied and modified
// here, so here is the project's copyright notice.

/*
MIT License

Copyright (c) 2017-2018 the project authors.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

func unmarshalError(jsonUnmarshalErr error) error {
	if before, after, found := strings.Cut(jsonUnmarshalErr.Error(), "Bis"); found && before != "" && after != "" {
		before = strings.ReplaceAll(before, " Go struct ", " ")
		return fmt.Errorf("%s%s", before, strings.ReplaceAll(after, "Bis", ""))
	}
	return jsonUnmarshalErr
}

type Parameter struct {
	Extensions       map[string]any      `json:"-" yaml:"-"`
	Ref              string              `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	In               string              `json:"in,omitempty" yaml:"in,omitempty"`
	Name             string              `json:"name,omitempty" yaml:"name,omitempty"`
	Description      string              `json:"description,omitempty" yaml:"description,omitempty"`
	CollectionFormat string              `json:"collectionFormat,omitempty" yaml:"collectionFormat,omitempty"`
	Type             string              `json:"type,omitempty" yaml:"type,omitempty"`
	Format           string              `json:"format,omitempty" yaml:"format,omitempty"`
	Pattern          string              `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	AllowEmptyValue  bool                `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	Required         bool                `json:"required,omitempty" yaml:"required,omitempty"`
	UniqueItems      bool                `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
	ExclusiveMin     bool                `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
	ExclusiveMax     bool                `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
	Schema           *openapi2.SchemaRef `json:"schema,omitempty" yaml:"schema,omitempty"`
	Items            *openapi2.SchemaRef `json:"items,omitempty" yaml:"items,omitempty"`
	Enum             []any               `json:"enum,omitempty" yaml:"enum,omitempty"`
	MultipleOf       *float64            `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
	Minimum          *float64            `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	Maximum          *float64            `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	MaxLength        *uint64             `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	MaxItems         *uint64             `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	MinLength        uint64              `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MinItems         uint64              `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	Default          any                 `json:"default,omitempty" yaml:"default,omitempty"`
}

func (parameter *Parameter) UnmarshalJSON(data []byte) error {
	type ParameterBis Parameter
	var x ParameterBis

	if err := json.Unmarshal(data, &x); err != nil {
		return unmarshalError(err)
	}

	_ = json.Unmarshal(data, &x.Extensions)
	delete(x.Extensions, "$ref")

	delete(x.Extensions, "in")
	delete(x.Extensions, "name")
	delete(x.Extensions, "description")
	delete(x.Extensions, "collectionFormat")
	delete(x.Extensions, "type")
	delete(x.Extensions, "format")
	delete(x.Extensions, "pattern")
	delete(x.Extensions, "allowEmptyValue")
	delete(x.Extensions, "required")
	delete(x.Extensions, "uniqueItems")
	delete(x.Extensions, "exclusiveMinimum")
	delete(x.Extensions, "exclusiveMaximum")
	delete(x.Extensions, "schema")
	delete(x.Extensions, "items")
	delete(x.Extensions, "enum")
	delete(x.Extensions, "multipleOf")
	delete(x.Extensions, "minimum")
	delete(x.Extensions, "maximum")
	delete(x.Extensions, "maxLength")
	delete(x.Extensions, "maxItems")
	delete(x.Extensions, "minLength")
	delete(x.Extensions, "minItems")
	delete(x.Extensions, "default")

	if len(x.Extensions) == 0 {
		x.Extensions = nil
	}

	*parameter = Parameter(x)
	return nil
}

type Operation struct {
	Extensions   map[string]any                 `json:"-" yaml:"-"`
	Summary      string                         `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description  string                         `json:"description,omitempty" yaml:"description,omitempty"`
	Deprecated   bool                           `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	ExternalDocs *openapi3.ExternalDocs         `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Tags         []string                       `json:"tags,omitempty" yaml:"tags,omitempty"`
	OperationID  string                         `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters   []Parameter                    `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Responses    map[string]*openapi2.Response  `json:"responses" yaml:"responses"`
	Consumes     []string                       `json:"consumes,omitempty" yaml:"consumes,omitempty"`
	Produces     []string                       `json:"produces,omitempty" yaml:"produces,omitempty"`
	Schemes      []string                       `json:"schemes,omitempty" yaml:"schemes,omitempty"`
	Security     *openapi2.SecurityRequirements `json:"security,omitempty" yaml:"security,omitempty"`
}

// UnmarshalJSON sets Operation to a copy of data.
func (operation *Operation) UnmarshalJSON(data []byte) error {
	type OperationBis Operation
	var x OperationBis
	if err := json.Unmarshal(data, &x); err != nil {
		return unmarshalError(err)
	}
	_ = json.Unmarshal(data, &x.Extensions)
	delete(x.Extensions, "summary")
	delete(x.Extensions, "description")
	delete(x.Extensions, "deprecated")
	delete(x.Extensions, "externalDocs")
	delete(x.Extensions, "tags")
	delete(x.Extensions, "operationId")
	delete(x.Extensions, "parameters")
	delete(x.Extensions, "responses")
	delete(x.Extensions, "consumes")
	delete(x.Extensions, "produces")
	delete(x.Extensions, "schemes")
	delete(x.Extensions, "security")
	if len(x.Extensions) == 0 {
		x.Extensions = nil
	}
	*operation = Operation(x)
	return nil
}

type PathItem struct {
	Extensions map[string]any `json:"-" yaml:"-"`
	Delete     *Operation     `json:"delete,omitempty" yaml:"delete,omitempty"`
	Get        *Operation     `json:"get,omitempty" yaml:"get,omitempty"`
	Head       *Operation     `json:"head,omitempty" yaml:"head,omitempty"`
	Options    *Operation     `json:"options,omitempty" yaml:"options,omitempty"`
	Patch      *Operation     `json:"patch,omitempty" yaml:"patch,omitempty"`
	Post       *Operation     `json:"post,omitempty" yaml:"post,omitempty"`
	Put        *Operation     `json:"put,omitempty" yaml:"put,omitempty"`
	Parameters []Parameter    `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

// UnmarshalJSON sets PathItem to a copy of data.
func (pathItem *PathItem) UnmarshalJSON(data []byte) error {
	type PathItemBis PathItem
	var x PathItemBis
	if err := json.Unmarshal(data, &x); err != nil {
		return unmarshalError(err)
	}
	_ = json.Unmarshal(data, &x.Extensions)
	delete(x.Extensions, "$ref")
	delete(x.Extensions, "delete")
	delete(x.Extensions, "get")
	delete(x.Extensions, "head")
	delete(x.Extensions, "options")
	delete(x.Extensions, "patch")
	delete(x.Extensions, "post")
	delete(x.Extensions, "put")
	delete(x.Extensions, "parameters")
	if len(x.Extensions) == 0 {
		x.Extensions = nil
	}
	*pathItem = PathItem(x)
	return nil
}

type SwaggerDoc struct {
	Extensions          map[string]any                      `json:"-" yaml:"-"`
	Swagger             string                              `json:"swagger" yaml:"swagger"` // required
	Info                openapi3.Info                       `json:"info" yaml:"info"`       // required
	ExternalDocs        *openapi3.ExternalDocs              `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Schemes             []string                            `json:"schemes,omitempty" yaml:"schemes,omitempty"`
	Consumes            []string                            `json:"consumes,omitempty" yaml:"consumes,omitempty"`
	Produces            []string                            `json:"produces,omitempty" yaml:"produces,omitempty"`
	Host                string                              `json:"host,omitempty" yaml:"host,omitempty"`
	BasePath            string                              `json:"basePath,omitempty" yaml:"basePath,omitempty"`
	Paths               map[string]PathItem                 `json:"paths,omitempty" yaml:"paths,omitempty"`
	Definitions         map[string]*openapi2.SchemaRef      `json:"definitions,omitempty" yaml:"definitions,omitempty"`
	Parameters          map[string]*Parameter               `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Responses           map[string]*openapi2.Response       `json:"responses,omitempty" yaml:"responses,omitempty"`
	SecurityDefinitions map[string]*openapi2.SecurityScheme `json:"securityDefinitions,omitempty" yaml:"securityDefinitions,omitempty"`
	Security            openapi2.SecurityRequirements       `json:"security,omitempty" yaml:"security,omitempty"`
	Tags                openapi3.Tags                       `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// UnmarshalJSON sets T to a copy of data.
func (doc *SwaggerDoc) UnmarshalJSON(data []byte) error {
	type TBis SwaggerDoc
	var x TBis
	if err := json.Unmarshal(data, &x); err != nil {
		return unmarshalError(err)
	}
	_ = json.Unmarshal(data, &x.Extensions)
	delete(x.Extensions, "swagger")
	delete(x.Extensions, "info")
	delete(x.Extensions, "externalDocs")
	delete(x.Extensions, "schemes")
	delete(x.Extensions, "consumes")
	delete(x.Extensions, "produces")
	delete(x.Extensions, "host")
	delete(x.Extensions, "basePath")
	delete(x.Extensions, "paths")
	delete(x.Extensions, "definitions")
	delete(x.Extensions, "parameters")
	delete(x.Extensions, "responses")
	delete(x.Extensions, "securityDefinitions")
	delete(x.Extensions, "security")
	delete(x.Extensions, "tags")
	if len(x.Extensions) == 0 {
		x.Extensions = nil
	}
	*doc = SwaggerDoc(x)
	return nil
}

func createKinParameter(parameter *Parameter) *openapi2.Parameter {
	if parameter == nil {
		return nil
	}

	return &openapi2.Parameter{
		// This one attribute we need to correct, as we can't unmarshal to Type.
		Type: &openapi3.Types{parameter.Type},
		// This attributes all map 1:1
		Ref:              parameter.Ref,
		Extensions:       parameter.Extensions,
		In:               parameter.In,
		Name:             parameter.Name,
		Description:      parameter.Description,
		CollectionFormat: parameter.CollectionFormat,
		Format:           parameter.Format,
		Pattern:          parameter.Pattern,
		AllowEmptyValue:  parameter.AllowEmptyValue,
		Required:         parameter.Required,
		UniqueItems:      parameter.UniqueItems,
		ExclusiveMin:     parameter.ExclusiveMin,
		ExclusiveMax:     parameter.ExclusiveMax,
		Schema:           parameter.Schema,
		Items:            parameter.Items,
		Enum:             parameter.Enum,
		MultipleOf:       parameter.MultipleOf,
		Minimum:          parameter.Minimum,
		Maximum:          parameter.Maximum,
		MaxLength:        parameter.MaxLength,
		MaxItems:         parameter.MaxItems,
		MinLength:        parameter.MinLength,
		MinItems:         parameter.MinItems,
		Default:          parameter.Default,
	}
}

func createKinParameters(parameters []Parameter) []*openapi2.Parameter {
	kinParameters := make([]*openapi2.Parameter, len(parameters))

	for i, parameter := range parameters {
		kinParameters[i] = createKinParameter(&parameter)
	}

	return kinParameters
}

func createKinOperation(operation *Operation) *openapi2.Operation {
	if operation == nil {
		return nil
	}

	return &openapi2.Operation{
		// This attribute needs fixing.
		Parameters: createKinParameters(operation.Parameters),
		// These attributes map 1:1
		Extensions:   operation.Extensions,
		Summary:      operation.Summary,
		Description:  operation.Description,
		Deprecated:   operation.Deprecated,
		ExternalDocs: operation.ExternalDocs,
		Tags:         operation.Tags,
		OperationID:  operation.OperationID,
		Responses:    operation.Responses,
		Consumes:     operation.Consumes,
		Produces:     operation.Produces,
		Schemes:      operation.Schemes,
		Security:     operation.Security,
	}
}

// Unmarshal loads a YAML or JSON Swagger document into a kin-openapi representation.
func UnmarshalSwagger(data []byte, kinDoc *openapi2.T) error {
	var fixedDoc SwaggerDoc

	if err := json.Unmarshal(data, &fixedDoc); err != nil {
		return err
	}

	// Copy properties that map 1:1 to kin-openapi parameters.
	kinDoc.Extensions = fixedDoc.Extensions
	kinDoc.Swagger = fixedDoc.Swagger
	kinDoc.Info = fixedDoc.Info
	kinDoc.ExternalDocs = fixedDoc.ExternalDocs
	kinDoc.Schemes = fixedDoc.Schemes
	kinDoc.Consumes = fixedDoc.Consumes
	kinDoc.Produces = fixedDoc.Produces
	kinDoc.Host = fixedDoc.Host
	kinDoc.BasePath = fixedDoc.BasePath
	kinDoc.Definitions = fixedDoc.Definitions
	kinDoc.Responses = fixedDoc.Responses
	kinDoc.SecurityDefinitions = fixedDoc.SecurityDefinitions
	kinDoc.Security = fixedDoc.Security
	kinDoc.Tags = fixedDoc.Tags

	kinDoc.Parameters = make(map[string]*openapi2.Parameter)

	for key, parameter := range fixedDoc.Parameters {
		kinDoc.Parameters[key] = createKinParameter(parameter)
	}

	kinDoc.Paths = make(map[string]*openapi2.PathItem)

	for key, pathItem := range fixedDoc.Paths {
		kinDoc.Paths[key] = &openapi2.PathItem{
			// This attribute is fixed.
			Parameters: createKinParameters(pathItem.Parameters),
			// All operations are fixed so types in Parameters in them can be fixed.
			Delete:  createKinOperation(pathItem.Delete),
			Get:     createKinOperation(pathItem.Get),
			Head:    createKinOperation(pathItem.Head),
			Options: createKinOperation(pathItem.Options),
			Patch:   createKinOperation(pathItem.Patch),
			Post:    createKinOperation(pathItem.Post),
			Put:     createKinOperation(pathItem.Put),
			// These parameters all map 1:1.
			Extensions: pathItem.Extensions,
		}
	}

	return nil
}

// OpenAPI30ToSwagger converts OpenAPI 3.0 documents from kin-openapi with additional fixes applied.
func OpenAPI30ToSwagger(kinOpenAPIDoc *openapi3.T) (*openapi2.T, error) {
	// kin-openapi errors on request body content, so rip it all out.
	contentLocations := extractSwaggerRequestBodyContent(kinOpenAPIDoc)

	kinSwaggerDoc, err := openapi2conv.FromV3(kinOpenAPIDoc)

	if err != nil {
		return nil, err
	}

	// Put the request body content we ripped out back in again.
	insertOpenAPI30RequestBodyContent(kinSwaggerDoc, contentLocations)

	return kinSwaggerDoc, nil
}

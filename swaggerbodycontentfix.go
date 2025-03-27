package openapispecconverter

import (
	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi3"
)

type operationType int

const (
	operationOptions operationType = iota
	operationHead
	operationGet
	operationPost
	operationPatch
	operationPut
	operationDelete
)

type bodyContentLocation struct {
	path          string
	operationType operationType
}

type contentLocations map[bodyContentLocation]openapi3.Content

func extractOperationBodyContent(operation *openapi3.Operation) openapi3.Content {
	if operation != nil &&
		operation.RequestBody != nil &&
		operation.RequestBody.Value != nil &&
		len(operation.RequestBody.Value.Content) > 0 {
		content := operation.RequestBody.Value.Content
		operation.RequestBody.Value.Content = nil

		return content
	}

	return nil
}

func extractSwaggerRequestBodyContent(kinOpenAPIDoc *openapi3.T) contentLocations {
	extractedContent := make(contentLocations)

	for path, pathItem := range kinOpenAPIDoc.Paths.Map() {
		if content := extractOperationBodyContent(pathItem.Options); content != nil {
			extractedContent[bodyContentLocation{path: path, operationType: operationOptions}] = content
		}

		if content := extractOperationBodyContent(pathItem.Post); content != nil {
			extractedContent[bodyContentLocation{path: path, operationType: operationPost}] = content
		}

		if content := extractOperationBodyContent(pathItem.Patch); content != nil {
			extractedContent[bodyContentLocation{path: path, operationType: operationPatch}] = content
		}

		if content := extractOperationBodyContent(pathItem.Put); content != nil {
			extractedContent[bodyContentLocation{path: path, operationType: operationPut}] = content
		}
	}

	return extractedContent
}

func insertOpenAPI30RequestBodyContent(kinSwaggerDoc *openapi2.T, contentToAdd contentLocations) {
	for location, content := range contentToAdd {
		if pathItem, ok := kinSwaggerDoc.Paths[location.path]; ok {
			var operation *openapi2.Operation

			switch location.operationType {
			case operationOptions:
				operation = pathItem.Options
			case operationPatch:
				operation = pathItem.Patch
			case operationPost:
				operation = pathItem.Post
			case operationPut:
				operation = pathItem.Put
			}

			if mediaType := content.Get("application/json"); mediaType != nil && operation != nil {
				parameter := openapi2.Parameter{
					Extensions: mediaType.Extensions,
					Name:       "body",
					In:         "body",
					Required:   true,
				}

				if mediaType.Schema != nil {
					if len(mediaType.Schema.Ref) > 0 {
						parameter.Schema = &openapi2.SchemaRef{
							Ref: mediaType.Schema.Ref,
						}
					} else if mediaType.Schema.Value != nil {
						parameter.Schema = &openapi2.SchemaRef{
							Value: &openapi2.Schema{
								Type:            parameter.Type,
								Format:          parameter.Format,
								Enum:            parameter.Enum,
								Min:             parameter.Minimum,
								Max:             parameter.Maximum,
								ExclusiveMin:    parameter.ExclusiveMin,
								ExclusiveMax:    parameter.ExclusiveMax,
								MinLength:       parameter.MinLength,
								MaxLength:       parameter.MaxLength,
								Default:         parameter.Default,
								Items:           parameter.Items,
								MinItems:        parameter.MinItems,
								MaxItems:        parameter.MaxItems,
								Pattern:         parameter.Pattern,
								AllowEmptyValue: parameter.AllowEmptyValue,
								UniqueItems:     parameter.UniqueItems,
								MultipleOf:      parameter.MultipleOf,
							},
						}
					}
				}

				operation.Parameters = append(operation.Parameters, &parameter)
			}
		}
	}
}

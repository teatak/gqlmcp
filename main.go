package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/teatak/gqlmcp/common"
)

type Empty struct {
}

type GraphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

var endpoint = ""
var introspectionQuery = `
query IntrospectionQuery {
  __schema {
    queryType {
      name
      kind
    }
    mutationType {
      name
      kind
    }
    subscriptionType {
      name
      kind
    }
    types {
      ...FullType
    }
    directives {
      name
      description
      locations
      args {
        ...InputValue
      }
    }
  }
}

fragment FullType on __Type {
  kind
  name
  description
  fields(includeDeprecated: true) {
    name
    description
    args {
      ...InputValue
    }
    type {
      ...TypeRef
    }
    isDeprecated
    deprecationReason
  }
  inputFields {
    ...InputValue
  }
  interfaces {
    ...TypeRef
  }
  enumValues(includeDeprecated: true) {
    name
    description
    isDeprecated
    deprecationReason
  }
  possibleTypes {
    ...TypeRef
  }
}

fragment InputValue on __InputValue {
  name
  description
  type {
    ...TypeRef
  }
  defaultValue
}

fragment TypeRef on __Type {
  kind
  name
  ofType {
    kind
    name
    ofType {
      kind
      name
      ofType {
        kind
        name
        ofType {
          kind
          name
          ofType {
            kind
            name
            ofType {
              kind
              name
              ofType {
                kind
                name
                ofType {
                  kind
                  name
                  ofType {
                    kind
                    name
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}
		`

func main() {
	endpoint = os.Getenv("URL")
	if endpoint == "" {
		endpoint = "https://countries.trevorblades.com/"
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "mcp-graphql",
		Version: common.Version,
	}, nil)

	server.AddResource(&mcp.Resource{
		Name:        "graphql-schema",
		Description: "access graphql schema",
		URI:         endpoint,
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		result, err := executeGraphQL(introspectionQuery, nil)
		if err != nil {
			return nil, err
		}
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      req.Params.URI,
					MIMEType: "application/json",
					Text:     string(result),
				},
			},
		}, nil
	})

	server.AddTool(&mcp.Tool{
		Name:        "introspect_schema",
		Description: "introspect the GraphQL schema, use this tool before doing a query to get the schema information if you do not have it available as a resource already.",
		InputSchema: &jsonschema.Schema{
			Type: "object",
		},
	},
		func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			result, err := executeGraphQL(introspectionQuery, nil)
			if err != nil {
				return nil, err
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: result},
				},
			}, nil
		},
	)

	server.AddTool(
		&mcp.Tool{
			Name:        "graphql_request",
			Description: "query a GraphQL endpoint with the given query and variables",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"query": {
						Type:        "string",
						Description: "the GraphQL query string (e.g. 'query { user { name } }')",
					},
					"variables": {
						Type:        "object",
						Description: "optional variables",
					},
				},
				Required: []string{"query"},
			},
		},
		func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := &GraphQLRequest{}
			if err := json.Unmarshal(req.Params.Arguments, args); err != nil {
				return nil, err
			}
			result, err := executeGraphQL(args.Query, args.Variables)
			if err != nil {
				return nil, err
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: result},
				},
			}, nil
		},
	)
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

func executeGraphQL(query string, variables map[string]interface{}) (string, error) {
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if headersEnv := os.Getenv("HEADERS"); headersEnv != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(headersEnv), &headers); err == nil {
			for k, v := range headers {
				req.Header.Set(k, v)
			}
		} else {
			parts := strings.SplitN(headersEnv, ":", 2)
			if len(parts) == 2 {
				req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		}
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// We treat HTTP 4xx/5xx as errors, but GraphQL errors (inside 200 OK) are returned as content.
	if resp.StatusCode >= 400 {
		return string(respBody), fmt.Errorf("http status %d", resp.StatusCode)
	}

	return string(respBody), nil
}

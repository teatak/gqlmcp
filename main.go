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

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

func main() {

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "mcp-graphql",
		Version: common.Version,
	}, nil)

	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "inspect_schema",
			Description: "Introspect the GraphQL schema, use this tool before doing a query to get the schema information if you do not have it available as a resource already.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
			introspectionQuery := `
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
			result, err := executeGraphQL(introspectionQuery, nil)
			if err != nil {
				return nil, nil, err
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: result},
				},
			}, nil, nil
		},
	)

	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "graphql_request",
			Description: "Query a GraphQL endpoint with the given query and variables",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"query": {
						Type:        "string",
						Description: "GraphQL query string (e.g., 'query { user { name } }')",
					},
					"variables": {
						Type:        "object",
						Description: "Optional variables JSON object",
					},
				},
				Required: []string{"query"},
			}},
		func(ctx context.Context, req *mcp.CallToolRequest, args GraphQLRequest) (*mcp.CallToolResult, any, error) {
			result, err := executeGraphQL(args.Query, args.Variables)
			if err != nil {
				return nil, nil, err
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: result},
				},
			}, nil, nil
		},
	)
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

func executeGraphQL(query string, variables map[string]interface{}) (string, error) {

	endpoint := os.Getenv("URL")
	if endpoint == "" {
		endpoint = "https://countries.trevorblades.com/"
		log.Printf("URL env var not set, using default demo URL: %s", endpoint)
	}

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

	if headerStr := os.Getenv("HEADER"); headerStr != "" {
		parts := strings.SplitN(headerStr, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			req.Header.Set(key, value)
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

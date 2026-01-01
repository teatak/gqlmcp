# gqlmcp

[中文文档](./README_CN.md)

An [MCP (Model Context Protocol)](https://github.com/modelcontextprotocol/model-context-protocol) server that acts as a proxy for GraphQL APIs. This allows MCP clients (like Claude Desktop) to introspect schemas and execute GraphQL queries against any GraphQL endpoint.

## Features

- **Schema Introspection**: Exposes the GraphQL schema as an MCP resource (`graphql://schema`) and a tool (`introspect_schema`).
- **Query Execution**: Provides a tool (`graphql_request`) to execute arbitrary GraphQL queries and mutations.
- **Header Support**: configure custom headers (e.g., for authentication) via environment variables.

## Usage

### Prerequisites

- Go 1.25+ (for building from source)
- An MCP Client (e.g., Claude Desktop, Zed)

### Installation

#### Build from Source

```bash
# Clone the repository
git clone https://github.com/teatak/gqlmcp.git
cd gqlmcp

# Build for your platform (Linux/macOS/Windows)
make
# Or manually:
# go build -o gqlmcp main.go
```

The binary will be located in the `bin/` or `pkg/` directory.

### Configuration

Add the server to your MCP client configuration (e.g., `claude_desktop_config.json`).

#### Environment Variables

- `URL`: **(Required)** The URL of the GraphQL API endpoint. (Default: `https://countries.trevorblades.com/`)
- `HEADERS`: **(Optional)** Custom headers for the request. Can be a JSON object or a `Key: Value` string.
  - Example (JSON): `'{"Authorization": "Bearer token123", "X-Custom-Header": "value"}'`
  - Example (Simple): `'Authorization: Bearer token123'`

#### Example Configuration (Claude Desktop)

```json
{
  "mcpServers": {
    "countries-graphql": {
      "command": "/path/to/gqlmcp",
      "env": {
        "URL": "https://countries.trevorblades.com/"
      }
    },
    "my-private-api": {
      "command": "/path/to/gqlmcp",
      "env": {
        "URL": "https://api.example.com/graphql",
        "HEADERS": "{\"Authorization\": \"Bearer my-secret-token\"}"
      }
    }
  }
}
```

## Tools Available

1.  **`introspect_schema`**: Returns the full GraphQL schema introspection result. Useful for the AI to understand the available types and queries.
2.  **`graphql_request`**: Executes a GraphQL query.
    -   Arguments:
        -   `query` (string): The GraphQL query.
        -   `variables` (object, optional): Query variables.

## Resources

-   **`graphql://schema`**: Direct access to the schema introspection result as a resource.

## License

[MIT](LICENSE)

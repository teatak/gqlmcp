# gqlmcp

mcp graphql proxy

config 
```yaml
name: booking
version: 0.0.1
schema: v1
mcpServers:
  - name: booking
    command: ./gqlmcp/gqlmcp
    env: 
      URL: "https://countries.trevorblades.com/"
      HEADER: '{"Authorization": "Token mytoken"}'
```
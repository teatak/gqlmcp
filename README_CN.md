# gqlmcp

一个 [MCP (Model Context Protocol)](https://github.com/modelcontextprotocol/model-context-protocol) 服务器，作为 GraphQL API 的代理。它允许 MCP 客户端（如 Claude Desktop）对任意 GraphQL 端点进行 Schema 内省并执行查询。

## 功能特性

- **Schema 内省**：将 GraphQL schema 暴露为 MCP 资源 (`graphql://schema`) 和工具 (`introspect_schema`)。
- **查询执行**：提供工具 (`graphql_request`) 来执行任意 GraphQL 查询和变更 (Mutations)。
- **Header 支持**：通过环境变量配置自定义请求头（例如用于身份验证）。

## 使用指南

### 前置要求

- Go 1.25+ (如果从源码构建)
- 一个 MCP 客户端 (例如 Claude Desktop, Zed)

### 安装

#### 源码构建

```bash
# 克隆仓库
git clone https://github.com/teatak/gqlmcp.git
cd gqlmcp

# 为你的平台构建 (Linux/macOS/Windows)
make
# 或者手动构建:
# go build -o gqlmcp main.go
```

构建后的二进制文件将位于 `bin/` 或 `pkg/` 目录中。

### 配置

将服务器添加到你的 MCP 客户端配置文件中（例如 `claude_desktop_config.json`）。

#### 环境变量

- `URL`：**(必填)** GraphQL API 的端点地址。(默认值: `https://countries.trevorblades.com/`)
- `HEADERS`：**(可选)** 请求的自定义 Header。可以是 JSON 对象或 `Key: Value` 字符串格式。
  - 示例 (JSON): `'{"Authorization": "Bearer token123", "X-Custom-Header": "value"}'`
  - 示例 (简单): `'Authorization: Bearer token123'`

#### 配置示例 (Claude Desktop)

```json
{
  "mcpServers": {
    "countries-graphql": {
      "command": "/绝对路径/指向/gqlmcp",
      "env": {
        "URL": "https://countries.trevorblades.com/"
      }
    },
    "my-private-api": {
      "command": "/绝对路径/指向/gqlmcp",
      "env": {
        "URL": "https://api.example.com/graphql",
        "HEADERS": "{\"Authorization\": \"Bearer my-secret-token\"}"
      }
    }
  }
}
```

## 可用工具

1.  **`introspect_schema`**：返回完整的 GraphQL schema 内省结果。这有助于 AI 理解可用的类型和查询字段。
2.  **`graphql_request`**：执行 GraphQL 查询。
    -   参数：
        -   `query` (string): GraphQL 查询语句。
        -   `variables` (object, optional): 查询变量。

## 资源

-   **`graphql://schema`**：作为资源直接访问 schema 内省结果。

## 许可证

[MIT](LICENSE)

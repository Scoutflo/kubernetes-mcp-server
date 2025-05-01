# Kubernetes MCP Server

[✨ Features](#features) | [🚀 Getting Started](#getting-started) | [🎥 Demos](#demos) | [⚙️ Configuration](#configuration) | [🧑‍💻 Development](#development)


## ✨ Features <a id="features"></a>

A powerful and flexible Kubernetes [Model Context Protocol (MCP)](https://blog.marcnuri.com/model-context-protocol-mcp-introduction) server implementation with support for **Kubernetes** .

- **✅ Configuration**:
  - Automatically detect changes in the Kubernetes configuration and update the MCP server.
  - **View** and manage the current [Kubernetes `.kube/config`](https://blog.marcnuri.com/where-is-my-default-kubeconfig-file) or in-cluster configuration.
- **✅ Generic Kubernetes Resources**: Perform operations on **any** Kubernetes.
  - Any CRUD operation (Create or Update, Get, List, Delete).
- **✅ Pods**: Perform Pod-specific operations.
  - **List** pods in all namespaces or in a specific namespace.
  - **Get** a pod by name from the specified namespace.
  - **Delete** a pod by name from the specified namespace.
  - **Show logs** for a pod by name from the specified namespace.
  - **Exec** into a pod and run a command.
  - **Run** a container image in a pod and optionally expose it.
- **✅ Namespaces**: List Kubernetes Namespaces.
- **✅ Events**: View Kubernetes events in all namespaces or in a specific namespace.


Unlike other Kubernetes MCP server implementations, this **IS NOT** just a wrapper around `kubectl` or `helm` command-line tools.

There is **NO NEED** for external dependencies or tools to be installed on the system.
If you're using the native binaries you don't even need to have Node or Python installed on your system.

## 🚀 Getting Started <a id="getting-started"></a>

### Requirements

- Access to a Kubernetes cluster.

### Claude Desktop

#### Using npx

If you have npm installed, this is the fastest way to get started with `kubernetes-mcp-server` on Claude Desktop and cursor.

Open your `claude_desktop_config.json` and add the mcp server to the list of `mcpServers`:
``` json
{
  "mcpServers": {
    "kubernetes": {
      "command": "npx",
      "args": [
        "-y",
        "kubernetes-mcp-server@latest"
      ]
    }
  }
}
```

Or using binary:

```json
{
  "mcpServers": {
    "kubernetes": {
      "command": "npx",
      "args": [
        "-y",
        "/path/to/binary/kubernetes-mcp-server"
      ]
    }
  }
}
```

### VS Code / VS Code Insiders

Install the Kubernetes MCP server extension in VS Code Insiders by pressing the following link:

[<img alt="Install in VS Code Insiders" src="https://img.shields.io/badge/VS_Code_Insiders-VS_Code_Insiders?style=flat-square&label=Install%20Server&color=24bfa5">](https://insiders.vscode.dev/redirect?url=vscode-insiders%3Amcp%2Finstall%3F%257B%2522name%2522%253A%2522kubernetes%2522%252C%2522command%2522%253A%2522npx%2522%252C%2522args%2522%253A%255B%2522-y%2522%252C%2522kubernetes-mcp-server%2540latest%2522%255D%257D)

Alternatively, you can install the extension manually by running the following command:

```shell
# For VS Code
code --add-mcp '{"name":"kubernetes","command":"npx","args":["kubernetes-mcp-server@latest"]}'
# For VS Code Insiders
code-insiders --add-mcp '{"name":"kubernetes","command":"npx","args":["kubernetes-mcp-server@latest"]}'
```

### Goose CLI

[Goose CLI](https://blog.marcnuri.com/goose-on-machine-ai-agent-cli-introduction) is the easiest (and cheapest) way to get rolling with artificial intelligence (AI) agents.

#### Using npm

If you have npm installed, this is the fastest way to get started with `kubernetes-mcp-server`.

Open your goose `config.yaml` and add the mcp server to the list of `mcpServers`:
```yaml
extensions:
  kubernetes:
    command: npx
    args:
      - -y
      - kubernetes-mcp-server@latest

```

## ⚙️ Configuration <a id="configuration"></a>

The Kubernetes MCP server can be configured using command line (CLI) arguments.

```shell
# Run the Kubernetes MCP server using npx (in case you have npm installed)
npx kubernetes-mcp-server@latest --help
```

```shell
# Run the Kubernetes MCP server using the latest release binary
./kubernetes-mcp-server --help
```

### Configuration Options

| Option        | Description                                                                                                                                                                                                                                                                                   |
|---------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `--sse-port`  | Starts the MCP server in Server-Sent Event (SSE) mode and listens on the specified port.                                                                                                                                                                                                      |
| `--log-level` | Sets the logging level (values [from 0-9](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/logging.md)). Similar to [kubectl logging levels](https://kubernetes.io/docs/reference/kubectl/quick-reference/#kubectl-output-verbosity-and-debugging). |

## 🧑‍💻 Development <a id="development"></a>

### Running with mcp-inspector

Compile the project and run the Kubernetes MCP server with [mcp-inspector](https://modelcontextprotocol.io/docs/tools/inspector) to inspect the MCP server.

```shell
# Compile the project
make build
# Run the Kubernetes MCP server with mcp-inspector
npx @modelcontextprotocol/inspector@latest $(pwd)/kubernetes-mcp-server
```



## Image naming convention for k8s-mcp-server
Docker image: `scoutflo/kubernetes_mcp_server`

## Tagging Strategy
- Pre-production (release/dev branch):
  - pre-prod-v{major}.{minor} (e.g., pre-prod-v1.0, pre-prod-v1.1)
  - pre-prod-{run_number} (build number tracking)
- Production (main branch):
  - prod-v{major}.{minor} (e.g., prod-v1.0, prod-v1.1)
  - production-{run_number} (build number tracking)

## Pipeline Image Upgrade Logic
The workflow fetches the latest Docker tags
For pre-prod (release/dev branch):
- Extracts major and minor versions from latest pre-prod-v{major}.{minor} tag
- Increments the minor version: pre-prod-v{major}.{minor+1}
For production (master branch):
- Extracts major and minor from latest prod-v{major}.{minor} tag
- Increments the minor version: prod-v{major}.{minor+1}
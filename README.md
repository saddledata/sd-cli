# 🛰️ sd: The Saddle Data CLI

`sd` is the official command-line interface for Saddle Data. It allows data engineers and developers to manage, monitor, and deploy data pipelines and AI validation contracts directly from the terminal.

## 🚀 Key Features

- **Declarative Infrastructure:** Apply your data stack using YAML specs (`sd apply`).
- **Operational Visibility:** Stream logs and monitor flow health (`sd flow monitor`).
- **AI Reliability:** Interactively test AI payload extraction and schema coercion (`sd gateway validate`).
- **Integrated Governance:** Run `pii-hound` scans locally (`sd scan`).
- **Context Management:** Seamlessly switch between dev, staging, and prod environments.

---

## 📥 Installation

### From Source (Go)

```bash
cd sd-cli
go build -o sd main.go
# Move it to your path
mv sd /usr/local/bin/
```

---

## 🛠️ Getting Started

### 1. Authenticate

Authenticate your CLI using an API Key generated from the Saddle Data Governance Control Center.

```bash
sd auth login --key your_api_key_here
```

### 2. Set Up Contexts

By default, `sd` uses `https://api.saddledata.io`. You can add other environments (like local development):

```bash
# Add a dev context with a self-signed cert
sd context set dev --api-url https://localhost:8080 --insecure

# Switch to it
sd context use dev
```

---

## 📂 Infrastructure as Code (IaC)

Apply your configuration idempotently:

```bash
sd apply -f saddle.yaml
```

---

## 🌊 Flow Management

### List Flows
```bash
sd flow list
```

### Trigger a Sync
```bash
sd flow run <flow-id>
```

### Stream Live Logs
```bash
sd flow monitor <flow-id>
# OR
sd flow logs <flow-id> --follow
```

---

## 🤖 AI Reliability (Gateway)

Test your LLM validation contracts:

```bash
sd gateway validate --asset <asset-id> --file raw_llm_output.txt
```

---

## 🛡️ Governance & Security

Scan local files or databases for PII using the integrated `pii-hound` engine:

```bash
sd scan ./data/*.csv
sd scan postgres://localhost:5432/db
```

---

## ⚙️ Configuration (`~/.sd.yaml`)

The CLI stores its configuration in `~/.sd.yaml`. You can manually edit this file to manage contexts and keys.

### Example Structure:
```yaml
active_context: default
contexts:
    default:
        api_url: https://api.saddledata.io
        api_key: sd_prod_xxx
    dev:
        api_url: https://localhost:8080
        api_key: sd_local_xxx
        insecure_skip_verify: true
```

| Field | Description |
|-------|-------------|
| `active_context` | The name of the currently active environment. |
| `contexts` | A map of environment configurations. |
| `api_url` | The base URL of the Saddle Data API for that context. |
| `api_key` | The API Key used for authentication in that context. |
| `insecure_skip_verify` | (Optional) If `true`, the CLI will skip TLS certificate verification. |

---

## 🛡️ Maintainers

The `sd` CLI is maintained by the team at Saddle Data.

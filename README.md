# 📊 Insti Attend GSheets

[![CI](https://img.shields.io/github/actions/workflow/status/arghyadipchak/insti-attend-gsheets/ci.yml?logo=github-actions&logoColor=white)](https://github.com/arghyadipchak/insti-attend-gsheets/actions/workflows/ci.yml)
[![GitHub Release](https://img.shields.io/github/v/release/arghyadipchak/insti-attend-gsheets?logo=github&logoColor=white&color=0969da)](https://github.com/arghyadipchak/insti-attend-gsheets/releases/latest)
[![Docker Image Version](https://img.shields.io/docker/v/arghyadipchak/insti-attend-gsheets?sort=semver&logo=docker&logoColor=white&color=2496ed)](https://hub.docker.com/r/arghyadipchak/insti-attend-gsheets)
[![Go Version](https://img.shields.io/github/go-mod/go-version/arghyadipchak/insti-attend-gsheets?logo=go&logoColor=white&color=00add8)](go.mod)
[![Commitizen Friendly](https://img.shields.io/badge/commitizen-friendly-brightgreen?logo=git&logoColor=white)](http://commitizen.github.io/cz-cli/)
[![License](https://img.shields.io/github/license/arghyadipchak/insti-attend-gsheets?color=8250df)](LICENSE)

A lightweight, [PocketBase](https://pocketbase.io/)-powered service that synchronizes attendance logs into Google Sheets through authenticated webhooks and background cron workers.

Built as the Google Sheets sync companion for [**Insti Attend**](https://github.com/arghyadipchak/insti-attend)

---

## 📑 Table of Contents

- [⚙️ Environment Variables](#️-environment-variables)
- [🚀 Hosting & Deployment](#-hosting--deployment)
  - [🐳 Option 1: Docker & Docker Compose](#-option-1-docker--docker-compose)
  - [📦 Option 2: Standalone Binary (Releases)](#-option-2-standalone-binary-releases)
- [🛠️ Development](#️-development)
  - [Local Dev Server](#local-dev-server)
  - [Build & Test](#build--test)
- [📡 Webhook API Reference](#-webhook-api-reference)
  - [Endpoint](#endpoint)
  - [Authentication](#authentication)
  - [Request Payload](#request-payload)
  - [Response Codes](#response-codes)
  - [Example `curl` Request](#example-curl-request)
- [📄 License](#-license)

---

## ⚙️ Environment Variables

| Variable | Description | Default |
| :--- | :--- | :--- |
| `USER_EMAIL` | Superuser email created on first startup | *None* |
| `USER_PASSWORD` | Superuser password created on first startup | *None* |
| `APP_URL` | Public application URL | *None* |
| `CRON_SYNC` | Cron expression for background attendance sync | `* * * * *` (every minute) |
| `TZ` | Timezone identifier (e.g., `Asia/Kolkata`, `UTC`) | System default / `UTC` |

> 💡 **CLI Flags**: Server bind address and data directory are passed as flags to the binary:
> - `--http=127.0.0.1:8090` (bind address and port, default: `127.0.0.1:8090`)
> - `--dir=./data` (database and storage directory, default: `./pb_data`)

---

## 🚀 Hosting & Deployment

### 🐳 Option 1: Docker & Docker Compose

#### Using Docker Compose (Recommended)

1. Use the provided `compose.yml` (or create one):
   ```yaml
   services:
     iattend-gsheets:
       image: arghyadipchak/insti-attend-gsheets:latest
       container_name: iattend-gsheets
       restart: unless-stopped
       environment:
         USER_EMAIL: admin@example.com
         USER_PASSWORD: your_secure_password
         TZ: Asia/Kolkata
       volumes:
         - iattend-gsheets-data:/data
       ports:
         - 8090:8090

   volumes:
     iattend-gsheets-data:
       name: iattend-gsheets-data
   ```

2. Start the container in detached mode:
   ```bash
   docker compose up -d
   ```

3. View live logs:
   ```bash
   docker compose logs -f
   ```

#### Using Standalone Docker

```bash
# Pull and run the image with a persistent volume
docker run -d \
  --name iattend-gsheets \
  --restart unless-stopped \
  -p 8090:8090 \
  -e USER_EMAIL=admin@example.com \
  -e USER_PASSWORD=your_secure_password \
  -e TZ=Asia/Kolkata \
  -v iattend-gsheets-data:/data \
  arghyadipchak/insti-attend-gsheets:latest
```

---

### 📦 Option 2: Standalone Binary (Releases)

You can download prebuilt binaries from the GitHub Releases page or build an optimized release binary from source.

1. **Download & make executable**:
   ```bash
   chmod +x insti-attend-gsheets
   ```

2. **Run the binary**:
   ```bash
   export USER_EMAIL=admin@example.com
   export USER_PASSWORD=your_secure_password

   ./insti-attend-gsheets serve --http=0.0.0.0:8090 --dir=./data
   ```

---

## 🛠️ Development

### Local Dev Server

Start the development server with automatic migrations and local data storage:

```bash
make dev
# or: go run . serve --dir ./data
```

Or run using the local development container:

```bash
docker compose -f compose.dev.yml up --build
```

Admin UI: 👉 **`http://localhost:8090/_/`**

### Build & Test

```bash
# Run tests & lint checks
make check

# Build binary
make build    # or make release for release build

# View all available make targets
make help
```

---

## 📡 Webhook API Reference

### Endpoint

```http
POST /wbh/{slug}
```

- `{slug}`: Unique webhook identifier configured in the Admin UI

### Authentication

Pass the generated webhook secret as a Bearer token in the `Authorization` header:

```http
Authorization: Bearer <TOKEN_SECRET>
Content-Type: application/json
```

### Request Payload

Send a JSON object mapping roll numbers / student identifiers to their attendance records:

```json
{
  "21CS01001": {
    "timestamp": "2026-08-17T09:00:00+05:30"
  },
  "21CS01002": {
    "timestamp": "2026-08-17T09:02:15+05:30"
  },
  "21CS01003": {
    "timestamp": "2026-08-17T09:05:40+05:30"
  }
}
```

#### Field Details
- **`timestamp`** *(string, required)*: ISO 8601 / RFC 3339 formatted timestamp string

### Response Codes

| Status Code | Status | Meaning |
| :--- | :--- | :--- |
| **`201 Created`** | `{"id": "..."}` | Attendance batch received and queued for synchronization |
| **`204 No Content`** | *Empty* | Empty payload or no spreadsheets attached to this webhook |
| **`400 Bad Request`** | Error detail | Invalid JSON format or missing `timestamp` field |
| **`401 Unauthorized`** | Error detail | Missing, invalid, or unauthorized Bearer token |
| **`404 Not Found`** | Error detail | Webhook slug does not exist or webhook is disabled |

### Example `curl` Request

```bash
curl -X POST https://your-domain.com/wbh/class-cs101 \
  -H "Authorization: Bearer YOUR_64_CHAR_TOKEN_SECRET" \
  -H "Content-Type: application/json" \
  -d '{
    "21CS01001": {
      "timestamp": "2026-08-17T09:30:00+05:30"
    },
    "21CS01002": {
      "timestamp": "2026-08-17T09:30:45+05:30"
    }
  }'
```

---

## 📄 License

This project is licensed under the **GNU Affero General Public License v3.0** ([AGPL-3.0](LICENSE))

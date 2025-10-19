# OpenGoSSLChecker

OpenGoSSLChecker is a simple tool to check SSL certificates of services.

<p align="center">
<img src="https://raw.githubusercontent.com/leblanc-simon/open-go-ssl-checker/main/static/img/logo.png">
</p>

## Installation

### From source

```bash
git clone https://github.com/leblanc-simon/open-go-ssl-checker.git
cd open-go-ssl-checker
make release
```

### From binary

Download the binary from the [releases page](https://github.com/leblanc-simon/open-go-ssl-checker/releases).

## Usage

### Launch the tool

```bash
# Quick launch with default configuration
./open-go-ssl-checker

# Launch with custom configuration
./open-go-ssl-checker -config /path/to/config.yml
```

You can go to `http://127.0.0.1:4332` to access the web interface.

### Configuration

The configuration file is a YAML file. The default configuration file is located at `config.yml`.

You can indicate the path to the configuration file with the `-config` flag.

Example of configuration file:

```yaml
# Configuration file for OpenGoSSLChecker

# Database configuration
database:
  driver: sqlite3
  dsn: ./ogsc.db

# Server configuration
server:
  port: 4332
  host: 127.0.0.1
  log_level: error
  api_key: "change-me-please"  # Optional; if set, required for API access
```

You can also use environment variables to configure the tool. The environment variables are prefixed with `OGSC_`.

Example:

```bash
export OGSC_DB_DRIVER=sqlite3
export OGSC_DB_DSN=./ogsc.db
export OGSC_SERVER_PORT=4332
export OGSC_SERVER_HOST=127.0.0.1
export OGSC_LOG_LEVEL=error
export OGSC_API_KEY="change-me-please"
```

### API

An authenticated HTTP API is available to add projects.

- Endpoint: POST /api/projects
- Auth: send your API key in the X-API-Key header. The key is configured via server.api_key in the YAML config or OGSC_API_KEY env var.
- Content-Type: application/json
- Request body:
  {
    "name": "My Project",
    "host": "example.com",
    "port": 443,
    "type": "https",
    "allow_insecure": false
  }
- Responses:
  - 201 Created: { "id": "<uuid>", "status": "created" }
  - 400 Bad Request: { "error": "..." }
  - 401 Unauthorized: { "error": "unauthorized" }

Example:

```bash
curl -X POST http://127.0.0.1:4332/api/projects \
  -H "Content-Type: application/json" \
  -H "X-API-Key: change-me-please" \
  -d '{
    "name": "My Project",
    "host": "example.com",
    "port": 443,
    "type": "https",
    "allow_insecure": false
  }'
```

### Compilation

If you want cross-compile the tool for another architecture, make sure you have the right tools installed :

* Linux ARM64 : sudo apt-get install -y gcc-aarch64-linux-gnu g++-aarch64-linux-gnu
* Windows : sudo apt-get install -y mingw-w64

### Author

* Simon Leblanc <contact@leblanc-simon.eu>

## License

[WTFPL](http://www.wtfpl.net/)

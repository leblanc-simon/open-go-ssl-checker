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
```

You can also use environment variables to configure the tool. The environment variables are prefixed with `OGSC_`.

Example:

```bash
export OGSC_DB_DRIVER=sqlite3
export OGSC_DB_DSN=./ogsc.db
export OGSC_SERVER_PORT=4332
export OGSC_SERVER_HOST=127.0.0.1
export OGSC_LOG_LEVEL=error
```

### Author

* Simon Leblanc <contact@leblanc-simon.eu>

## License

[WTFPL](http://www.wtfpl.net/)

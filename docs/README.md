# Caddy WACE

---

## Introduction



## Configuration
This section provides details on each field within the configuration files and their purposes. These configurations are necessary to control how the system processes requests, handles plugins, manages logging, and more.

### File: waceconfig.yaml
- logpath (String): Specifies the path to the log file where all WACE related logs will be stored. Default is typically set to /var/log/wace.log.
- loglevel (String): Sets the log level for the application. Possible values include DEBUG, INFO, WARN or ERROR.

**Plugins configuration**
- modelplugins: Defines the plugins used for processing transactions.
  - id (String): a unique identifier for each model plugin.
  - plugintype (String): Specifies the plugin type. Possible values include "RequestHeaders", "RequestBody", "AllRequest", "ResponseHeaders", "ResponseBody", and "AllResponse".
  - path (String): file path to the plugin executable file.
  - weight (Float): defines the weight of this plugin in scoring decisions.
  - mode (String): execution mode, values can be "sync" or "async".
  - remote (Boolean): indicates whether the plugin is executed through NATS.

- decisionplugins: contains plugins used to determine final actions based on model plugin outputs.
  - id (String): identifier for each decision plugin.
  - path (String): file path to the plugin executable file.
  - params: Contains parameters for decision-making logic.
    - waf_weight (String): weight assigned to Web Application Firewall (WAF) in decision scoring.
    - threshold (String): minimum threshold score to apply the decision plugin’s result.

**Network and Options**

- natsurl (String): URL for the NATS server, which handles messaging between components. Default format is hostname:port.
- options:
  - otelurl (String): URL for the OpenTelemetry collector in order to send metrics.
  - crs_version (String): version of the OWASP CRS in use.
  - early_blocking (String): enables or disables early blocking of requests (true or false).
  - listenaddress (String), optional: IP address to bind the grpc server.
  - listenport (String), optional: port that grpc server listen. 

### File: <app>waceappconfig.yaml
- modelids (List): a list of model plugin IDs to be applied in this application. Each ID SHOULD match a modelplugin defined in the main configuration.
- decisionid (String): specifies the ID of the decision plugin to use. This ID should correspond to one in the decisionplugins section of the main configuration.
- options:
  - appname (String): the name of the application. This can be used for identification purposes.

### File: waceexceptions.conf

This file defines models exception using SecLang syntax. In order to except a particular model, you need to add in your rule the following directive:
  ```bash
  setvar:tx.<model_id>=false
  ```
By default, all models defined in the configuration are applied to each transaction.

## Installing with Docker Compose

To install the project with [Docker Compose](https://docs.docker.com/compose/install/), follow these steps:

   ```bash
   git clone https://gitlab.fing.edu.uy/gsi/pgrado-wace/caddy_wace.git
   cd caddy_wace
   docker compose up -d
   ```

WordPress is available at http://localhost:8080. 

You can visualize the metrics in the Grafana dashboard. The default credentials are admin/admin. The dashboard is available at http://localhost:3000.

## Manual installation

If you plan to install the project manually, you need to accomplish the following requirements:

**Requirements:**
- **Programming Language**: [Go](https://go.dev/doc/install)

**Optional:**

- If you plan to use Async Models, you need to deploy a [NATS server](https://docs.nats.io/running-a-nats-service/introduction/installation).
- In order to export metrics, you need to deploy an [OpenTelemetry Collector](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/examples/otel-collector).

## Steps for manual installation

First of all, if you plan to use NATS or OpenTelemetry services, please install it before follow the next steps.

1. Clone this repository:

   ```bash
   https://gitlab.fing.edu.uy/gsi/pgrado-wace/caddy_wace.git
   cd caddy_wace
   ```
2. Build your plugins. We provide an example for trivial plugins here.
   ```bash
   go build -buildmode=plugin -o ./_plugins/trivial.so ./_plugins/trivial.go 
   go build -buildmode=plugin -o ./_plugins/trivial2.so ./_plugins/trivial2.go 
   go build -buildmode=plugin -o ./_plugins/weighted_sum.so ./_plugins/weighted_sum.go 
   ```
3. Build it.
   ```bash
   go build .
   ```

### Manual usage

You can configure Caddy deployment in the *Caddyfile*. Then, just run with caddy directives. For example:
   ```bash
   ./caddy run
   ```
## Contributing

Merge requests are welcome. For major changes, please open an issue first
to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

# config-db

**config-db** is developer first, JSON based configuration management database (CMDB).

## Setup

### Setup local db link as environment variable.

```bash
export DB_URL=postgres://<username>:<password>@localhost:5432/config
```

### Create `config` database.

```sql
create database config
```

### Scape config and serve

Starting the server will run the migrations and start scraping in background (The `default-schedule` configuration will run scraping every 60 minutes if configuration is not explicitly specified).

```bash
make build

./.bin/config-db serve
```

To explicitly run scraping with a particular configuration:

```bash
./.bin/config-db run <scrapper-config.yaml> -vvv
config-db serve
```

See `fixtures/` for example scraping configurations.

### Migrations

Commands `./bin/config-db serve` or `./bin/config-db run` would run the migrations.

Setup [goose](https://github.com/pressly/goose) for more options on migration. Goose commands need to be run from `db/migrations` directory.

```bash
GOOSE_DRIVER=postgres GOOSE_DBSTRING="user=postgres dbname=config sslmode=disable" goose down
```
## Principles

* **JSON Based** - Configuration is stored in JSON, with changes recorded as JSON patches that enables highly structured search.
* **SPAM Free** - Not all configuration data is useful, and overly verbose change histories are difficult to navigate.
* **GitOps Ready** - Configuration should be stored in Git, config-db enables the extraction of configuration out of Git repositories with branch/environment awareness.
* **Topology Aware** - Configuration can often have an inheritance or override hierarchy.

## Capabilities

* View and search change history in any dimension (node, zone, environment, application, technology)
* Compare and diff configuration across environments.

## Configuration Sources

* AWS
  * [x] EC2 (including trusted advisor, compliance and patch reporting)
  * [x] VPC
  * [ ] IAM
* Kubernetes
  * [ ] Pods
  * [ ] Secrets / ConfigMaps
  * [ ] LoadBalancers / Ingress
  * [ ] Nodes
* Configuration Files
  * [ ] YAML/JSON
  * [ ] Properties files
* Dependency Graphs
  * [ ] pom.xml
  * [ ] package.json
  * [ ] go.mod
* Infrastructure as Code
  * [ ] Terraform
  * [ ] CloudFormation
  * [ ] Ansible

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md)

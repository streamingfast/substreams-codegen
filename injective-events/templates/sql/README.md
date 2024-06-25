# Description

- This is a generated Substreams with bindings to write to an SQL database

# Dependencies

## Get Substreams-sink-sql

* Get latest release from https://github.com/streamingfast/substreams-sink-sql

## Get Substreams CLI (optional)

To try the Substreams directly, you need to install the `substreams CLI` (v1.7.2 or above).

You have many options as explained in this [installation guide](https://substreams.streamingfast.io/documentation/consume/installing-the-cli).

Check if `substreams` was installed successfully, you can run the following command:

```bash
substreams --version
```

## Get Substreams API Token

To stream data, you will need to get a Substreams API token.
Follow the instructions on the [authentification section](https://substreams.streamingfast.io/documentation/consume/authentication) in the `StreamingFast` documentation.

## Run the entire stack with the `run-local.sh` script

### Get Docker-compose

* following the instructions on the [official Docker website](https://docs.docker.com/get-docker/).

### Run the script 

This will launch a docker-compose environment with an `SQL database` the `substreams-sink-sql` service

```bash
./run-local.sh
```
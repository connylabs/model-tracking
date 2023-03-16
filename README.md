# model-tracking

`model-tracking` is a tool for storing tracking the results of machine learning services.

## Running model-tracking

### Prerequisites

`model-tracking` requires a PostgreSQL database in which to store results.
During development, this database can be provided using a PostgreSQL Docker container:

```shell
docker run --rm -it -p 5432:5432 -e POSTGRES_USER=model-tracking -e POSTGRES_PASSWORD=model-tracking -e POSTGRES_DB=model-tracking postgres
```

To prepare the database execute the following Makefile targets:

```shell
# Run the migrations against the database.
# Note: the database parameters can be overwritten by specifying the DATABASE_URL environment variable.
make migrate
```

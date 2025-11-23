---
title: One PostgreSQL container, multiple databases
description:
    Article talking about how to create multiple databases within one PostgreSQL container
date: "2025 November 23"
format: blog_entry
require_prism: true
---

# The need

I have multiple services that I develop in a single monorepo.

Almost all of these services need PostgreSQL to work and for clarity, I usually prefer having one database per service.

This is not a problem if you use your system's PostgreSQL but since I use Docker Compose for local development my PostgreSQL is a container using the official [postgres](https://hub.docker.com/_/postgres) image.
That image does _not_ support creating multiple databases by default.

Having one _container_ per service is overkill so I searched for another way to do that.

# The "standard" way

The official image supports adding [initialization scripts](https://hub.docker.com/_/postgres#initialization-scripts) that are run by the entrypoint which is enough for what I want.

The following is the script I'm using:
```bash
#!/bin/bash

set -e
set -u

function create_db_and_user() {
    local db_name=$1
    echo "Processing database and user '$db_name'"

    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "postgres" <<-EOSQL
        CREATE ROLE $db_name WITH LOGIN PASSWORD '$db_name' NOSUPERUSER NOCREATEDB IF NOT EXISTS;
        CREATE DATABASE $db_name OWNER $db_name IF NOT EXISTS;
        GRANT ALL PRIVILEGES ON DATABASE $db_name TO $db_name;
EOSQL
}

if [ -n "$POSTGRES_MULTIPLE_DATABASES" ]; then
    echo "Multiple database creation requested: $POSTGRES_MULTIPLE_DATABASES"

    # Split the comma-separated variable into an array and iterate
    for db in $(echo $POSTGRES_MULTIPLE_DATABASES | tr ',' ' '); do
        create_db_and_user $db
    done

    echo "Multiple databases processed"
fi
```

Put this in a file named `postgres-init-multiple-dbs.sh` and make sure it's executable with `chmod +x postgres-init-multiple-dbs.sh`.

Now you can start the container like this:
```bash
docker run -d --name postgres \
  -v ./postgres-init-multiple-dbs.sh:/docker-entrypoint-initdb.d/init-multiple-dbs.sh:ro \
  -e POSTGRES_USER=foo \
  -e POSTGRES_PASSWORD=foo \
  -e POSTGRES_MULTIPLE_DATABASES=foo,bar,baz \
  postgres:17-alpine
```

This will create the following users and databases: `foo`, `bar` and `baz`.

# With Docker Compose

Here's how I use this script in my `compose.yaml` file:

```yaml
postgres:
  image: postgres:17-alpine
  environment:
    POSTGRES_USER: foo
    POSTGRES_PASSWORD: foo
    POSTGRES_MULTIPLE_DATABASES: foo,bar,baz
  ports:
    - "5432:5432"
  volumes:
    - postgres_data:/var/lib/postgresql/data
    - ./postgres-init-multiple-dbs.sh:/docker-entrypoint-initdb.d/init-multiple-dbs.sh:ro
  healthcheck:
    test: ["CMD-SHELL", "pg_isready -U foo -d postgres"]
    interval: 10s
    timeout: 5s
    retries: 5
```

This works well and does not require creating a custom PostgreSQL image. Problem solved, right?

# With Github Actions

Not quite. In addition to Compose I'm also using a Github Actions workflow for my CI and the PostgreSQL requirements haven't changed.

Initially, I thought about using the [jobs.<job_id>.services.<service_id>.volumes](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax#jobsjob_idservicesservice_idvolumes) feature,
just like with my `compose.yaml` file. However, this didn't work as I expected, I kept running into this error from the postgres container:

```
/usr/local/bin/docker-entrypoint.sh: line 185: /docker-entrypoint-initdb.d/init-multiple-dbs.sh: Is a directory
```

Turns out, the service containers are started by the runner before the code is even checked out, so they can never access the script.

The solution is simple though: build a custom PostgreSQL image. The following `Dockerfile` does the job:
```dockerfile
FROM mirror.gcr.io/postgres:17-alpine

COPY postgres-init-multiple-dbs.sh /docker-entrypoint-initdb.d/init-multiple-dbs.sh
RUN chmod +x /docker-entrypoint-initdb.d/init-multiple-dbs.sh
```

Build and publish this image on a container registry, and you won't have to mess with mounting the script as a volume anywhere, simplifying everything.

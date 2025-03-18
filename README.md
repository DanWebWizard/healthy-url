
# Healthy URL Checker

This project checks the health of URLs concurrently and stores the results in a Cassandra database.

## Prerequisites

- Docker
- Docker Compose

## Setup

1. Clone the repository:
   ```sh
   git clone https://github.com/DanWebWizard/healthy-url.git
   cd healthy-url
   ```

2. Create a `.env` file in the root directory with the following content:
   ```properties
   # Number of worker threads
   WORKERS=10

   # Cassandra keyspace
   CASSANDRA_KEYSPACE=health_url

   # Cassandra host
   CASSANDRA_HOST=cassandra
   ```

3. Create a `urls.json` file in the root directory with the URLs you want to check. Example:
   ```json
   [
       "https://jsonplaceholder.typicode.com/posts",
       "https://jsonplaceholder.typicode.com/comments",
       "https://jsonplaceholder.typicode.com/albums",
       "https://jsonplaceholder.typicode.com/photos",
       "https://jsonplaceholder.typicode.com/todos",
       "https://jsonplaceholder.typicode.com/users",
       "https://dog.ceo/api/breeds/image/random",
       "https://dog.ceo/api/breeds/list/all",
       "https://dog.ceo/api/breed/hound/images",
       "https://dog.ceo/api/breed/hound/list",
       "https://catfact.ninja/fact",
       "https://catfact.ninja/breeds",
       "https://catfact.ninja/facts",
       "https://catfact.ninja/facts?limit=5",
       "https://catfact.ninja/facts?limit=10"
   ]
   ```

## Running the Project

1. Start the services using Docker Compose:
   ```sh
   docker-compose up
   ```

   This will build and start the `healthy-url` service and the `cassandra` service.

2. The `healthy-url` service will read the URLs from `urls.json`, check their health, and store the results in the Cassandra database.

## Viewing URL Results in Cassandra

1. Open a terminal and run the following command to access the Cassandra container:
   ```sh
   docker exec -it cassandra cqlsh
   ```

2. Once inside the Cassandra shell, use the keyspace specified in the `.env` file:
   ```sh
   USE health_url;
   ```

3. Query the `url_healths` table to view the results:
   ```sh
   SELECT * FROM url_healths;
   ```

## Environment Variables

- `WORKERS`: Number of worker threads to use for concurrent URL health checks. Default is 5 if not set.
- `CASSANDRA_KEYSPACE`: The keyspace to use in Cassandra. Default is `health_url` if not set.
- `CASSANDRA_HOST`: The host name of the Cassandra service. Default is `cassandra` if not set.

## Notes

- Ensure that the `urls.json` file is in the root directory of the project.
- The `.env` file should contain the necessary environment variables as shown above.
- The `docker-compose.yml` file defines the services and their configurations.

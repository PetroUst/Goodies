# Goodies

To run the project, you need to have the following installed:

- [Go](https://golang.org/)
- [Docker-desktop](https://www.docker.com/products/docker-desktop)
## Running the project

From the root of the project, run the following command:

```bash     
docker-compose up
```

This will start the server and other services. The server will be available at `http://localhost:8080`.

Possible endpoints are:

- `get?domain=<domain>` - Get the domain information
- `post` - Add or update url by domain in redis, mongo and mysql (body should be in json format like `{"domain": "example.com", "url": "https://example.com"}`)
- `suspicious` - Get random suspicious url from grpc server



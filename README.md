# Telemetry and Middleware example app

This is example of using [twistingmercury/middleware](https://github.com/twistingmercury/middleware) along with [twistingmercury/telemetry](https://github.com/twistingmercury/telemetry). The solution contains two go projects; a [client](./client) and a [server](./server).

## Server

 The server exposes a simple RESTful API using [gin & gonic](https://github.com/gin-gonic/gin), demonstrating how to use the middleware.
   ```go
    //...
    if err := middleware.Initialize(metrics.Registry(), namespace, serviceName); err != nil {
    logging.Fatal(err, "failed to initialize server middleware")
    }
    router := gin.New()
    router.Use(gin.Recovery(), middleware.Telemetry())
	// ...
   ```
The server's logging level can be adjusted using the environmental variable `LOG_LEVEL` in the [docker-compose.yaml](./docker-compose.yaml#L35). The default logging level is `warn`.

## Client

The client consumes this API via several goroutines to simulate several concurrent clients. This value can adjusted using the environmental variable `CONCURRENCY` in the [docker-compose.yaml](./docker-compose.yaml#L19).

The logging level of the client can be adjusted using the environmental variable `LOG_LEVEL` in the [docker-compose.yaml](./docker-compose.yaml#18). The default logging level is `warn`.

## Running the example

### Prerequisites

* Docker, Docker Compose, or a FOSS alternative such as [Rancher Desktop](https://rancherdesktop.io/)
* Go version >= 1.22

This example uses Jaeger for tracing, and Grafana for visualizing the metrics. You'll need basic knowledge on how to use these tools. These will be pulled automatically if you don't already have them locally.

### Running the example

* Simply execute `make` from the root directory. The solution will then build and containerize the server and the client, invoke the `docker compose up` command lauching the example.
* From there, to see the traces open [jaeger](http://localhost:16686). You should see something similar to this:

   ![image](assets/Screenshot%202024-06-25%20at%2012.27.35.png)

* To see the metrics, open [grafana](http://localhost:3000/explore/metrics/trail?from=now-1h&to=now&var-ds=P1809F7CD0C75ACF3&var-filters=&refresh=5s)
  * You'll need to log in. The user name and the password are both `admin`.
  * You'll be asked to change the password. You can skip this.

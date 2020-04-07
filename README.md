# go-load-balancer

Load balancer written with Go to experiment with concurrency models

## Reference

This load balancer was built following [this](https://kasvith.me/posts/lets-create-a-simple-lb-go/) tutorial.

## Installation

To install on your local machine, use the following commands.

```bash
git clone https://woojiahao/github.com/go-load-balancer.git
go build .
./go-load-balancer
```

## CLI 

### --backends

Loads the list of backends to load balance. This flag is **mandatory**. The list of backends must be comma-separated and the 
URLs cannot contain port numbers.

#### Usage

```bash
./go-load-balancer --backends 192.0.0.1,192.0.0.2
```

### --port

Changes the port of the load balancer. Defaults to **3030**. This flag is **optional**.

#### Usage

```bash
./go-load-balancer --port 8083
```


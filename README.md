# Kafka-Siege

### This tool provides a modular interface to implement low-level networking primitives for a Kafka client
Eg: TCP connectors; SSL connectors

### This tool is primarily used for modelling misbehaving clients 
Eg: Incomplete handshakes, Connection storms

## Build
- Harness
  ```bash
  go build -o ksiege cmd/kafkasiegecli/main.go 
  ```
- Plugin (Eg: TCP)
  ```bash
   go build -buildmode=plugin -o tcp.so plugins/tcp/tcp.go
  ```
  
## Run
- Configuration (as specified in `config.toml`)
- ```bash
  ./ksiege -config config.toml
  ```
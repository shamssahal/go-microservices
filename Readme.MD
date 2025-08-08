# Toll-Calculator

## Data Receiver

Context Hierarchy

```
signal.NotifyContext (main)
    └── ServerHTTP context
            └── handleWS context (per connection)
                    └── wsReceiveLoop context
```

## Install Protobuf compilers (protoc compiler)

```
brew install protobuff
```

```
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

```
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

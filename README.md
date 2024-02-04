# BrewKit

----

Container-native build system. Meaning that BrewKit is a **single tool** that you should use(across Docker) to build project.

BrewKit focuses on repeatable builds, host agnostic and build process containerization

----

## Key features

* [BuildKit](https://github.com/moby/buildkit) as core of build system. There is following features from BuildKit 
  * Distributed cache - inherited from BuildKit
  * Automatic garbage collection - inherited from BuildKit
* Aggressive-caching
* Mounted secrets
* Host configuration agnostic
* Output artifacts to Host filesystem
* JSON based build-definition format  
* [JSONNET](https://jsonnet.org/) configuration language - BrewKit uses jsonnet build to compile `brewkit.jsonnet` into JSON build-definition

## Naming

`BrewKit` - common style
<br/>
`brewkit` - go-style

## Start with BrewKit

Install BrewKit via go >= 1.20 
```shell
go install github.com/ispringtech/brewkit/cmd/brewkit
```

Create `brewkit.jsonnet`
```shell
touch brewkit.jsonnet
```

Describe simple target
```jsonnet
local app = "service";

local copy = std.native('copy');

{
    apiVersion: "brewkit/v1",
    targets: {
        all: ['gobuild'],
        
        gobuild: {
            from: "golang:1.20",
            workdir: "/app",
            copy: [
                copy('cmd', 'cmd'),
                copy('pkg', 'pkg'),
            ],
            command: std.format("go build -o ./bin/%s ./cmd/%s", [app])
        }
    }
}
```

Run build
```shell
brewkit build

 => resolve image config for docker.io/golangci/golangci-lint:v1.53                                                                                                                                         0.0s
 => resolve image config for docker.io/library/golang:1.20                                                                                                                                                  0.0s
 => [internal] Loading context                                                                                                                                                                              0.1s
 => => transferring build-context: 71.42kB                                                                                                                                                                  0.1s
 => CACHED docker-image://docker.io/library/golang:1.20                                                                                                                                                     0.0s
 => CACHED docker-image://docker.io/golangci/golangci-lint:v1.53                                                                                                                                            0.0s
 => _gosources                                                                                                                                                                                              0.2s
 => _gobase                                                                                                                                                                                                 0.1s
 => lint                                                                                                                                                                                                    2.7s
 => modules                                                                                                                                                                                                 1.0s
 => build                                                                                                                                                                                                   3.3s
 => test                                                                                                                                                                                                    1.7s
 => exporting to client directory                                                                                                                                                                           0.2s 
 => => copying files 30.83MB
# ...
```

## Build BrewKit

When brewkit installed locally
```shell
brewkit build
```

Build from source:
```shell
go build -o ./bin/brewkit ./cmd/brewkit
```

## Documentation

* [Documentaion entrypoint](docs/readme.md)
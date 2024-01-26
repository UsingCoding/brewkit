local images = import 'images.libsonnet';

local cache = std.native('cache');
local copy = std.native('copy');
local copyFrom = std.native('copyFrom');

// External cache for go compiler, go mod, golangci-lint
local gocache = [
    cache("go-build", "/app/cache"),
    cache("go-mod", "/go/pkg/mod"),
];

// Sources which will be tracked for changes
local gosources = [
    "go.mod",
    "go.sum",
    "cmd",
    "internal",
];

{
    project():: {
        apiVersion: "brewkit/v2",

        vars: {
            gitcommit: {
                from: images.golang,
                workdir: "/app",
                copy: copy('.git', '.git'),
                command: "git -c log.showsignature=false show -s --format=%H:%ct"
            }
        },

        targets: {
            all: ["build", "test", "lint"],

            build: {
                from: "_gobase",
                cache: gocache,
                copy: [
                    // copy go.mod changes
                    copyFrom(
                        'modules',
                        '/app/go.*',
                        '.'
                    )
                ],
                ssh: {},
                command: 'go build -trimpath -v -ldflags "-X main.Commit=${gitcommit}" -o ./bin/brewkit ./cmd/brewkit',
                output: "/app/bin/brewkit:./bin/",
            },

            test: {
                from: "_gobase",
                cache: gocache,
                dependsOn: ["modules"],
                command: "go test ./..."
            },

            modules: {
                from: "_gobase",
                copy: copyFrom(
                    '_gosources',
                    '/app',
                    '/app'
                ),
                cache: gocache,
                ssh: {},
                command: "go mod tidy",
                output: "/app/go.*:.",
            },

            lint: {
                from: images.golangcilint,
                workdir: "/app",
                cache: gocache,
                copy: [
                    copyFrom(
                        '_gosources',
                        '/app',
                        '/app'
                    ),
                    copy('.golangci.yml', '.golangci.yml'),
                ],
                env: {
                    GOCACHE: "/app/cache/go-build",
                    GOLANGCI_LINT_CACHE: "/app/cache/go-build"
                },
                command: "golangci-lint run"
            },

            // contains project go sources
            _gosources: {
                from: "scratch",
                workdir: "/app",
                copy: [copy(source, source) for source in gosources]
            },

            // base stage for all go targets
            _gobase: {
                from: "golang:1.20",
                workdir: "/app",
                env: {
                    GOCACHE: "/app/cache/go-build",
                },
                copy: copyFrom(
                    '_gosources',
                    '/app',
                    '/app'
                ),
            },
        }
    }
}
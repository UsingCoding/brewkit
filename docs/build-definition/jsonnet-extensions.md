# Jsonnet extensions

## cache

Allows to write cache definition as one-liner

See [build-definition reference](reference.md#cache)

```jsonnet
local cache = std.native('cache');
//
    targets: {
        gobuild: {
            // ...            
            cache: cache("go-build", "/app/cache"),
            // ...            
        }
    }
//    
```

## copy

Allows to write copy definition as one-liner

See [build-definition reference](reference.md#copy)

```jsonnet
local copy = std.native('copy');
//
    targets: {
        gobuild: {
            // ...            
            copy: [
                copy('cmd', 'cmd'),
                copy('pkg', 'pkg'),
            ],
            // ...            
        }
    }
//    
```

## copyFrom

Allows to write copyFrom definition as one-liner

See [build-definition reference](reference.md#copy)

```jsonnet
local copyFrom = std.native('copyFrom');
//
    targets: {
        gobuild: {
            // ...            
            copy: [
                copyFrom('prebuild', 'artifacts', 'artifacts'),
                copy('cmd', 'cmd'),
                copy('pkg', 'pkg'),
            ],
            // ...            
        }
    }
//    
```

## secret

Allows to write secret definition as one-liner

See [build-definition reference](reference.md#secrets)

```jsonnet
local secret = std.native('secret');
//
    targets: {
        gobuild: {
            // ...            
            secret: secret("aws", '/root/.aws/credentials'),
            // ...            
        }
    }
//    
```

## defArgSet

Allow to check if defArg is set

```jsonnet
local defArgSet = std.native('defArgSet');

if defArgSet("version") then {versionExists: true} else {}
```

## defArg

Returns defArg value. Returns empty string on non-existing argument

```jsonnet
local defArgSet = std.native('defArgSet');
local defArg = std.native('defArg');

if defArgSet("version") then {version: defArg("version")} else {}
```

Or simplified:

```jsonnet
local defArg = std.native('defArg');

if defArg("version") != "" then {version: defArg("version")} else {}
```

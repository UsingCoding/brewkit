# BrewKit host config

Config used to configure brewkit host preferences

## Location

By default, brewkit trying to find config in `${HOME}/.brewkit/config` or in  `$BREWKIT_CONFIG`. If there is no such path, used empty config. 
There is no config auto creation


## Reference

[Schema v1](/data/specification/config/v1.json) 

### Secrets

Define secret to use in build-definition. See [secrets in build-definition](/docs/build-definition/reference.md#secrets)

```jsonnet
{
    "secrets": [
        {
            // unique id for secret            
            "id": "aws",
            // path may contain env variables            
            "path": "${HOME}/.aws/credentials"
        },
    ]
}
```

### Entitlements

Allow brewkit some of the entitlements from buildkit.
For instance - use network host `network.host`

```jsonnet
{
    "entitlements": [
        "network.host"
    ]
}
```

Full list of available entitlements in config schema as enum (see [Schema v1](/data/specification/config/v1.json))

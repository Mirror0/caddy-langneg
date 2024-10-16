# Caddy Language Selector via Content Negotiation Plugin

## IMPORTANT: This is cut-down version of original plugin which only support `Accept-Language` header part of Content Negotiation and return language or full locale into a variable. For plugin with full support of Content Negotiation please use original author plugin.

[Content negotiation](https://en.wikipedia.org/wiki/Content_negotiation) is a mechanism of HTTP that allows client and server to agree on the best version of a resource to be delivered for the client's needs given the server's capabilities (see [RFC](https://datatracker.ietf.org/doc/html/rfc7231#section-5.3)). In short, when sending the request, the client can specify what *content type*, *language*, *character set* or *encoding* it prefers and the server responds with the best available version to fit the request.

This plugin to the [Caddy 2](https://caddyserver.com/) webserver allows you to configure [named matchers](https://caddyserver.com/docs/caddyfile/matchers#named-matchers) for content negotiation parameters and/or store content negotiation results in [variables](https://caddyserver.com/docs/conventions#placeholders).

The plugin can be configured via Caddyfile:

## Syntax

```Caddyfile
@name {
    langneg {
        match_languages <language codes...>
        full_locale <boolean>
        var_language <name>
        fallback_value <value>
    }
}
```

* `match_languages` takes one or more (space-separated) languages code (eg. en, de, en-US) that are available in this matcher. If the client requests a language (via HTTP's `Accept:` request header) compatible with one of those, the matcher returns true, if the request specifies types that cannot be satisfied by this list of offered types, the matcher returns false.
* `full_locale` is a boolean value that indicates that matcher should put full or closest to full locale information into `var_language` variable. 
* `var_language` allows you to define a string that, prefixed with `langneg_`, specifies a variable name that will store the result of the content type negotiation, i.e. the best content type according to the types and weights specified by the client and what is on offer by the server. You can access this variable with `{vars.langneg_<name>}` in other places of your configuration.
* `fallback_value` this is value will be used as-is as fallback if matcher does not match any value. In that case named matcher still will return true (required for caddy to execute named matcher) and provide this value in `langneg_<var_language>` variable if it was set.
* Requirements in the same named matcher are AND'ed together. If you want to OR, i.e. match alternatively, just configure multiple named matchers.
* You must specify `match_languages`. And when you specify one of the `var_language` parameter, `match_languages` parameter must be defined as well.
* Wildcards like `*` and `*/*` should work. If they don't behave as you expect, please open an issue.

A [Caddyfile](./Caddyfile) with some combinations for testing is provided with this repository. You can test it with commands like these:

```shell
$ curl -H "Accept-Language: fr-FR" https://localhost/
French
$ curl -H "Accept-Language: en" https://localhost/
German or English. en preferred!
$ curl -H "Accept-Language: en;q=0.4, de;q=0.8" https://localhost/
German or English. de preferred!
$ curl -H "Accept-Language: de-DE" https://localhost/
RDF auf deutsch oder englisch, de-DE preferred!
German. Full locale de-DE
$ curl -H "Accept-Language: en-US" https://localhost/
English. Short locale en
$ curl -H "Accept-Language: en-AU" https://localhost/
Fallback value english
$ curl https://localhost/
Default
```

## Libraries

The plugin relies heavily on go's own [x/text/language](https://pkg.go.dev/golang.org/x/text/language) libraries. (For the intricacies of language negotiation, you may want to have a glance at the [blog post](https://go.dev/blog/matchlang) that accompanied the release of go's language library.).

## License

This software is licensed under the Apache License, Version 2.0.

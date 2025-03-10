# Template Functions

- [Template Functions](#template-functions)
  - [`lower`, `lowercase`](#lower-lowercase)
  - [`upper`, `uppercase`](#upper-uppercase)
  - [`title`](#title)
  - [`sprintf`, `printf`](#sprintf-printf)
  - [`trim`](#trim)
  - [`trimPrefix`, `trimSuffix`](#trimprefix-trimsuffix)
  - [`default`](#default)
  - [`required`](#required)
  - [`env`](#env)
  - [`sha1sum`, `sha256sum`](#sha1sum-sha256sum)
  - [`str`](#str)
  - [`replace`](#replace)
  - [`alphanumify`, `alphanumdash`](#alphanumify-alphanumdash)
  - [`dottodash`, `dottounder`](#dottodash-dottounder)
  - [`index`](#index)

The following template functions are available, with some functions having aliases for convenience:

## `lower`, `lowercase`

Converts the value to string as stated in [String conversion](docs/faq.md#string-conversion), then lowercases it.

```handlebars
{{ "Namespace" | lower }}
namespace
```

## `upper`, `uppercase`

Converts the value to string as stated in [String conversion](docs/faq.md#string-conversion), then uppercases it.

```handlebars
{{ "Namespace" | upper }}
NAMESPACE
```

## `title`

Converts the value to string as stated in [String conversion](docs/faq.md#string-conversion), then capitalize the first character of each word.

```handlebars
{{ "hello world" | title }}
Hello World
```

While available, it's use is discouraged for file names.

## `sprintf`, `printf`

Alias of Go's `fmt.Sprintf`.

```handlebars
{{ printf "number-%d" 20 }}
number-20
```

## `trim`

Converts the value to string as stated in [String conversion](docs/faq.md#string-conversion), then removes any whitespace at the beginning or end of the string.

```handlebars
{{ "   hello world    " | trim }}
hello world
```

## `trimPrefix`, `trimSuffix`

Converts the value to string as stated in [String conversion](docs/faq.md#string-conversion), then removes either the prefix or the suffix.

Do note that the parameters are flipped from Go's `strings.TrimPrefix` and `strings.TrimSuffix`: here, the first parameter is the prefix, rather than being the last parameter. This is to allow piping one output to another:

```handlebars
{{ "   foo" | trimPrefix " " }}
foo
```

## `default`

If the value is set, return it, otherwise, a default value is used.

```handlebars
{{ "" | default "bar" }}
bar
```

## `required`

If the argument renders to an empty string, the application fails and exits with non-zero status code.

```handlebars
{{ "" | required }}
<!-- argument is marked as required, but it was not found in the YAML data -->
```

## `env`

Fetch an environment variable to be printed. If the environment variable is mandatory, consider using `required`. If the environment variable might be empty, consider using `default`.

`env` allows the key to be case-insensitive: it will be uppercased internally.

```handlebars
{{ env "user" }}
patrick
```

## `sha1sum`, `sha256sum`

Renders a `sha1sum` or `sha256sum` of a given value. The value is converted first to their YAML representation, with comments removed, then the `sum` is performed. This is to ensure that the "behavior" can stay the same, even when the file might have multiple comments that might change.

Primitives such as `string`, `bool` and `float64` are converted as-is.

While not recommended, you can use this to always generate a new name if the YAML declaration drifts. The following snippet uses `.`, which represents the entire YAML file -- on a multi-YAML file, each `.` represents a single file:

```handlebars
{{ . | sha1sum }}
f502bbf15d0988a9b28b73f8450de47f75179f5c
```

## `str`

Converts any primitive as stated in [String conversion](docs/faq.md#string-conversion), to string:

```handlebars
{{ false | str }}
false
```

## `replace`

Converts the value to a string as stated in [String conversion](docs/faq.md#string-conversion), then replaces all ocurrences of a string with another:

```handlebars
{{ "hello.dev" | replace "." "_" }}
hello_dev
```

## `alphanumify`, `alphanumdash`

Converts the value to a string as stated in [String conversion](docs/faq.md#string-conversion), and keeps from the original string only alphanumeric characters -- for `alphanumify` -- or alphanumeric plus dashes and underscores -- like URLs, for `alphanumdash`:

```handlebars
{{ "secret-foo.dev" | alphanumify }}
secretsfoodev
```

```handlebars
{{ "secret-foo.dev" | alphanumdash }}
secrets-foodev
```

## `dottodash`, `dottounder`

Converts the value to a string as stated in [String conversion](docs/faq.md#string-conversion), and replaces all dots to either dashes or underscores:

```handlebars
{{ "secret-foo.dev" | dottodash }}
secrets-foo-dev
```

```handlebars
{{ "secret-foo.dev" | dottounder }}
secrets-foo_dev
```

Particularly useful for Kubernetes FQDNs needed to be used as filenames.

## `index`

For certain resources where YAML indexes are not alphanumeric, but contain special characters such as labels or annotations, `index` allows you to retrieve those resources. Consider the following YAML:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
  labels:
    app.kubernetes.io/name: patrickdap-deployment
```

It's not possible to access the value `patrickdap-deployment` using dot notation like this: `{{ .metadata.labels.app.kubernetes.io/name }}`: the Go Template engine will throw an error. Instead, you can use `index`:

```handlebars
{{ index "app.kubernetes.io/name" .metadata.labels }}
patrickdap-deployment
```

The reason the parameters are flipped is to allow piping one output to another:

```handlebars
{{ .metadata.labels | index "app.kubernetes.io/name" }}
patrickdap-deployment
```

## `indexOrEmpty`

This function works the same as the index function but does not raise errors; instead, it returns an empty string, which can be used in if statements, piped to the default function, etc. For example:

```
{{ $component := indexOrEmpty "k8s.config/component" .metadata.labels | default "unlabeled" }}
 {{ printf "%s-%s-%s.yaml" $component (lower .kind) .metadata.name }}
```
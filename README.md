# trcache rueidis [![GoDoc](https://godoc.org/github.com/rrgmc/trcache-rueidis?status.png)](https://godoc.org/github.com/rrgmc/trcache-rueidis)

This is a [trcache](https://github.com/rrgmc/trcache) wrapper for [rueidis](https://github.com/rueian/rueidis).

## Info

### rueidis library

| info        |          |
|-------------|----------|
| Generics    | No       |
| Key types   | `string` |
| Value types | `string` |
| TTL         | Yes      |

### wrapper

| info              |                  |
|-------------------|------------------|
| Default codec     | `GOBCodec`       |
| Default key codec | `StringKeyCodec` |

## Installation

```shell
go get github.com/rrgmc/trcache-rueidis
```

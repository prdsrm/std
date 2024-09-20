# Contributing

Before creating pull requests, please read the [coding guidelines](https://github.com/uber-go/guide/blob/master/style.md).

TL;DR: KISS, and follow the UNIX philosophy.

General tradeoffs:

* Less is more
* Maintainability > feature bloat
* Simplicity > speed
* Consistency > elegance

Others:
- This library is set of **helpers**, and **higher-level** funtions for [td](https://github/gotd/td).
  Please avoid big wrappers.
  Focus on functions, that abstract the complex Telegram API for real-word usage, accept `td` objects,
  and return `td` objects as well, or data, directly in `[]byte`, or with multiple back-ends options.
- Avoid direct back-ends for the library functions. Don't save directly to a specific format, in it. Function should return data as `td` objects, or maybe, a widely used feed(such as WebSocket, for monitoring), with minimal modification of the data.
  Only commands and examples can export data to file, ideally, and if you really need to use a database like PostgreSQL it should probably be more than a simple command and a project on its own, in a separate repository.
- Avoid big commits, I prefer 10 small commits that lead to a single feature that 1 commit that lead
  to 10 features.
- Small, composable functions, separate as soon as they get too long. Same for packages.
I think you started to understand the idea...

## Coding guidance

Please read [Uber code style](https://github.com/uber-go/guide/blob/master/style.md).


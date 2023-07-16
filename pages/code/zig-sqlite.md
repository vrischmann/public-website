---
title: zig-sqlite - small wrapper around SQLite's C API
format: standard
require_prism: true
---

# zig-sqlite - small wrapper around SQLite's C API

[zig-sqlite](https://github.com/vrischmann/zig-sqlite) is a [Zig](https://ziglang.org/) wrapper around SQLite's C API.

# Introduction

[SQLite](https://sqlite.org/index.html) is a C library, directly usable in Zig using [`@cImport`](https://ziglang.org/documentation/0.8.0/#C).

While this works, the goal of `zig-sqlite` is to:
* abstract away the C semantics of directly using sqlite
* provide powerful, type safe abstractions for a _statement_ and _iterator_

# How it works

## C interop

`@cImport` parses the `sqlite3.h` header and import symbols such as functions, types, etc. Zig code can then use these symbols directly.

One problem with this is that sqlite is designed for C:
* errors are returned as an int
* strings are represented by a pointer to UTF-8 or UTF-16 bytes and a length integer

This doesn't match well with Zig's builtin features:
* [errors](https://ziglang.org/documentation/0.9.0/#Errors) is a first class language feature
* [slices](https://ziglang.org/documentation/0.9.0/#Slices) is also a first class language feature

Enabling my users to use these features is one of the primary goal of this wrapper.

## Type safe statement

I designed a _statement_ abstraction which make use of a _comptime_ known query and the `anytype` type.

The high level goal is to know at comptime the number and types of bind parameters, such as calls to _bind_ values will not compile if their types do not match.

This is achieved by parsing the query at comptime and extracting its bind parameters types.

The statement _type_ is then completely specific for this particular parsed query which allows us to type check the bind parameters.

The API looks something like this:

```zig
var stmt = try db.prepare("SELECT id FROM user WHERE name = ?{text} AND age < ?{usize}");
defer stmt.deinit();

const rows = try stmt.all(usize, .{
    .name = "Vincent",
    .age = @as(usize, 200),
});
defer rows.deinit();

for (rows) |row| {
    std.debug.print("row: {any}\n", .{row});
}
```

The statement obtained by calling `prepare` has a `all` method that will fail to compile if the provided bind parameters aren't of the correct type.

# Learn more

The [README](https://github.com/vrischmann/zig-sqlite) provides highly detailed information on how to use the library in your Zig code. Give it a try !

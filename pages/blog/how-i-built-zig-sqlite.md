---
title: How I built zig-sqlite
description: |
    Article explaining how to leverage Zig's compile time metaprogramming to build a SQLite wrapper that can type check SQL queries
date: "2022 May 26"
format: blog_entry
require_prism: true
---

# Introduction

I wrote and maintain [zig-sqlite](https://github.com/vrischmann/zig-sqlite), a Zig package that wraps the SQLite C API.

Its goals are to use Zig’s features to augment SQLite’s capabilities and provide a idiomatic Zig API.

# Why build a wrapper

Zig has [excellent C interoperability](https://ziglang.org/documentation/0.9.1/#C) which means using SQLite is just one [`@cInclude`](https://ziglang.org/documentation/0.9.1/#cInclude) away, however using the “raw” SQLite API has some drawbacks in the context of a Zig project:

- Zig has [slices](https://ziglang.org/documentation/0.9.1/#Slices), C does not. This means you need to pass the array or slice pointer and its length to the SQLite functions.
- Zig structs can have methods, C structs can’t.
- Zig has native [error sets](https://ziglang.org/documentation/0.9.1/#Errors), C doesn’t.

All of those issues can be fixed by introducing a Zig wrapper API.

In addition, using Zig’s [comptime](https://ziglang.org/documentation/0.9.1/#comptime) feature we can augment SQLite’s capabilities with:

- the ability to *type check* the bind parameters of a statement
- the ability to read a row of data into (almost) any type

This blog post will focus on how I used *comptime* to implement these features.

# What is `comptime`

Let’s start by seeing what `comptime` actually is.

`comptime` is Zig’s way of defining code that must evaluate to something known at *compile-time*.

You can do a lot with this:

- combined with types as first-class citizens, you can create new types from comptime code (this is how [generics](https://ziglang.org/documentation/0.9.1/#Generic-Data-Structures) are implemented)
- combined with [type reflection](https://ziglang.org/documentation/0.9.1/#typeInfo), you can create type-checked printf functions, JSON parser and more
- [comptime vars](https://ziglang.org/documentation/0.9.1/#Compile-Time-Variables) allow you to do complex computations at compile-time, for example you can imagine parsing some data at compile-time to generate a data structure which will later be used at runtime.

We use all of these capabilities in zig-sqlite.

# A brief demo of zig-sqlite

With all that in mind, this is how you use zig-sqlite today:

```zig
const std = @import("std");

const sqlite = @import("sqlite");

pub fn main() anyerror!void {
    var db = try sqlite.Db.init(.{
        .mode = sqlite.Db.Mode{ .Memory = {} },
        .open_flags = .{ .write = true },
    });
    defer db.deinit();

    try db.exec("CREATE TABLE user(id integer primary key, age integer, name text)", .{}, .{});

    const user_name: []const u8 = "Vincent";

    // Insert some data
    try db.exec("INSERT INTO user(id, age, name) VALUES($id{usize}, $age{u32}, $name{[]const u8})", .{}, .{
        @as(usize, 10),
        @as(u32, 34),
        user_name,
    });
    try db.exec("INSERT INTO user(id, age, name) VALUES($id{usize}, $age{u32}, $name{[]const u8})", .{}, .{
        @as(usize, 20),
        @as(u32, 84),
        @as([]const u8, "José"),
    });

    // Read one row into a struct
    const User = struct {
        id: usize,
        age: u32,
        name: []const u8,
    };
    const user_opt = try db.oneAlloc(User, std.testing.allocator, "SELECT id, age, name FROM user WHERE name = $name{[]const u8}", .{}, .{
        .name = user_name,
    });

    // Read single integers; reuse the same prepared statement
    var stmt = try db.prepare("SELECT id FROM user WHERE age = $age{u32}");
    defer stmt.deinit();

    const id1 = try stmt.one(usize, .{}, .{@as(u32, 34)});
    stmt.reset();
    const id2 = try stmt.one(usize, .{}, .{@as(u32, 84)});
}
```

The code is not complete, full demo is available [here](https://github.com/vrischmann/zig-sqlite-demo).

This showcases the features I want to talk about. I encourage you to play with the code (try to remove or change the type of a bind parameter).

# comptime-checked bind parameters

When you prepare a statement zig-sqlite creates a brand new type only for this prepared statement. This new type contains the *parsed query* which will be used to perform checks when you call either `iterator`, `exec`, `one` or `all` on a statement.

This implies the query is known at compile-time; in my experience and my projects this is almost always the case.

This is the trimmed code from zig-sqlite which creates this custom type (actual code [here](https://github.com/vrischmann/zig-sqlite/blob/b357fb1a6d799bea07e4e4c3972565d87c579340/sqlite.zig#L1498-L1510)):

```zig
pub fn StatementType(comptime query: []const u8) type {
    return Statement(ParsedQuery(query));
}

pub fn Statement(comptime query: anytype) type {
    return struct {
        // NOTE: lots of code here, omitted for clarity
    };
}

const query = "SELECT name FROM user WHERE age = $age{u32}";
var stmt: StatementType(query) = try db.prepare(query);
```

(If you’re wondering what `anytype` is, look [here](https://ziglang.org/documentation/0.9.1/#toc-Function-Parameter-Type-Inference)).

You can see how the statement type will be different for different queries.

When you call `exec`, `one` or `all` zig-sqlite will use this data on the statement to do two things:

- check that the number of bind parameters provided is strictly identical to the number of bind markers in the query
- check that each bind parameter has the correct type according to the *type annotation* of the bind marker

## Number of bind parameters

The first check is straightforward: executing a SQL query which has for example 4 bind markers with any other number of bind parameters is never correct.

With the query metadata in our statement type we can enforce this at compile-time.

Let’s see what happens if we provide the wrong number of bind parameters.

Given the following statement:

```zig
var stmt = try db.prepare("SELECT id FROM user WHERE age = $age{u32}");
```

Here’s what happens if we don’t respect the contract:

```txt
./third_party/zig-sqlite/sqlite.zig:2057:17: error: expected 1 bind parameters but got 2
                @compileError(comptime std.fmt.comptimePrint("expected {d} bind parameters but got {d}", .{
                ^
./third_party/zig-sqlite/sqlite.zig:2148:26: note: called from here
            try self.bind(.{}, values);
                         ^
./third_party/zig-sqlite/sqlite.zig:2194:41: note: called from here
            var iter = try self.iterator(Type, values);
                                        ^
./src/main.zig:43:29: note: called from here
    const id1 = try stmt.one(usize, .{}, .{ @as(u32, 34), @as(usize, 2000) });
                            ^
./src/main.zig:5:29: note: called from here
pub fn main() anyerror!void {
```

The error tells us that we expect exactly 1 bind parameter (because there is 1 bind marker) but we provided 2. We can also see directly in the error that we call `stmt.one` with a tuple of 2 values, which is the source of the compile error.

## Type of parameters

The second check is more involved and depends on a *type annotation*.

You might have already guessed with the demo code, zig-sqlite supports an “extended” version of SQL which allows a user to annotate a bind marker, instructing zig-sqlite to check that *all* values bound to this marker have this exact Zig type.

Let’s go back to the example above:

```zig
var stmt = try db.prepare("SELECT id FROM user WHERE age = $age{u32}");
```

This annotates the bind marker `$age` with the Zig type `u32`. Now anytime this statement is executed the `$age` bind parameter *must* have been bound to a `u32`; it’s a compilation error if that’s not the case.

Let’s see how it fails to compile when we try to pass a `u16` when a `u32` is expected:

```zig
_ = try stmt.one(usize, .{}, .{@as(u16, 34)});
```

Here is the error from the Zig compiler:

```txt
 ./third_party/zig-sqlite/sqlite.zig:2085:17: error: value type u16 is not the bind marker type u32
                @compileError("value type " ++ @typeName(Actual) ++ " is not the bind marker type " ++ @typeName(Expected));
                ^
./third_party/zig-sqlite/sqlite.zig:2072:58: note: called from here
                        else => comptime assertMarkerType(struct_field.field_type, typ),
                                                         ^
./third_party/zig-sqlite/sqlite.zig:2148:26: note: called from here
            try self.bind(.{}, values);
                         ^
./third_party/zig-sqlite/sqlite.zig:2194:41: note: called from here
            var iter = try self.iterator(Type, values);
                                        ^
./src/main.zig:43:29: note: called from here
    const id1 = try stmt.one(usize, .{}, .{@as(u16, 34)});
                            ^
./src/main.zig:5:29: note: called from here
pub fn main() anyerror!void {
```

The first line tells us exactly what’s wrong: we try to bind a value of type `u16` where a value of type `u32` is expected.

# Read a row of data into a type

Thanks to Zig’s *type reflection* we can read a row of data into a user-provided type without needing to write any “mapping” function: we know the type we want to read (here the `User` struct) and can analyse it at compile-time. With this data we can call the appropriate internal “read” functions.

## How it works

Let’s go back to the example in the demo:

```zig
const User = struct {
    id: usize,
    age: u32,
    name: []const u8,
};

const user_opt = try db.oneAlloc(User, std.testing.allocator, "SELECT id, age, name FROM user WHERE name = $name{[]const u8}", .{}, .{
    .name = user_name,
});
```

Notice we pass the type as first parameter. This type is ultimately used to create an [`Iterator`](https://github.com/vrischmann/zig-sqlite/blob/f32017ea460d00f0dd3ac134f64131a1dab93b44/sqlite.zig#L1033-L1054) which is responsible for reading data when [`next`](https://github.com/vrischmann/zig-sqlite/blob/f32017ea460d00f0dd3ac134f64131a1dab93b44/sqlite.zig#L1067) or [`nextAlloc`](https://github.com/vrischmann/zig-sqlite/blob/f32017ea460d00f0dd3ac134f64131a1dab93b44/sqlite.zig#L1125) is called.

So the meat of the code is in the iterator and it starts in [`nextAlloc`](https://github.com/vrischmann/zig-sqlite/blob/f32017ea460d00f0dd3ac134f64131a1dab93b44/sqlite.zig#L1124-L1207). The first step is to get the *type info* of the user type with [`@typeInfo`](https://ziglang.org/documentation/0.9.1/#typeInfo): this returns a value of type [`std.builtin.Type`](https://github.com/ziglang/zig/blob/326b2aa27bdf43b695798192e415db995ee9918b/lib/std/builtin.zig#L193) which is a tagged union we can analyse using a simple switch statement, for example:

```zig
switch (@typeInfo(UserType)) {
    .Int => processInt(),
    .Optional => processOptional(),
    .Array => processArray(),
    .Struct => processStruct(),
    else => @compileError("invalid type " ++ @typeName(UserType)),
}
```

Of course not all types make sense in the context of our `nextAlloc` function (it makes no sense to read into a `ComptimeInt` (which is the type of a comptime integer) or `Fn` (which is the type of a function) for example).

In our example we end up taking the [`.Struct`](https://github.com/vrischmann/zig-sqlite/blob/f32017ea460d00f0dd3ac134f64131a1dab93b44/sqlite.zig#L1199-L1204) prong which calls [`readStruct`](https://github.com/vrischmann/zig-sqlite/blob/f32017ea460d00f0dd3ac134f64131a1dab93b44/sqlite.zig#L1426-L1442). The `std.builtin.Type.Struct` type contains the struct fields which we iterate over *at compile-time* using a [`inline for`](https://ziglang.org/documentation/0.9.1/#inline-for).

Next we call [`readField`](https://github.com/vrischmann/zig-sqlite/blob/f32017ea460d00f0dd3ac134f64131a1dab93b44/sqlite.zig#L1444-L1494) for each struct field, giving it the field *type* and its position in the struct fields slice. `readField` also uses `@typeInfo` to do its thing and ultimately it will end up calling a specific read function depending on the field type.

In our example this would be:

- two calls to [`readInt`](https://github.com/vrischmann/zig-sqlite/blob/f32017ea460d00f0dd3ac134f64131a1dab93b44/sqlite.zig#L1256-L1260)
- one call to [`readBytes`](https://github.com/vrischmann/zig-sqlite/blob/f32017ea460d00f0dd3ac134f64131a1dab93b44/sqlite.zig#L1299-L1348)

Finally the result of the `readField` call is [assigned to the field](https://github.com/vrischmann/zig-sqlite/blob/f32017ea460d00f0dd3ac134f64131a1dab93b44/sqlite.zig#L1438) in the user type using [`@field`](https://ziglang.org/documentation/master/#field).

It’s important to remember that all these steps are done at *compile-time*; the code in the final binary would do something like this (not real code):

```zig
const user = User{
    .id = try stmt.readInt(usize, 1),
    .age = try stmt.readInt(u32, 2),
    .name = try stmt.readBytes([]const u8, std.testing.allocator, 3, .Text),
};
```

## Examples

I already demoed this before but here’s reading into a struct:

```zig
const User = struct {
    id: usize,
    age: u32,
    name: []const u8,
};

const user_opt = try db.oneAlloc(User, std.testing.allocator, "SELECT id, age, name FROM user WHERE name = $name{[]const u8}", .{}, .{
    .name = user_name,
});
```

You can also read a single type:

```zig
const age_opt = try db.one(usize, "SELECT age FROM user WHERE name = $name{[]const u8}", .{}, .{
    .name = user_name,
});
```

Note we’re not using the `oneAlloc` method here since we don’t need to allocate memory when reading a simple integer.

We also support reading into enums but this is a little more involved:

```zig
const Foo = enum(u7) {
    pub const BaseType = u16;

    low = 34,
    high = 84,
};

const foo_opt = try db.one(Foo, "SELECT age FROM user WHERE name = $name{[]const u8}", .{}, .{
    .name = user_name,
});
```

The `BaseType` type is mandatory and tells zig-sqlite how to read the *underlying* column from SQLite. Here we first read a `u16` value.

Next we convert the `u16` value to a `Foo` enum value.

We can also use the `BaseType` to store the enum value as a string:

```zig
const Foo = enum {
    pub const BaseType = []const u8;

    low,
    high,
};

const user_opt = try db.oneAlloc(Foo, std.testing.allocator, "SELECT name FROM user WHERE name = $name{[]const u8}", .{}, .{
    .name = user_name,
});
```

Now the stored value will be the enum *tag name* (here `low` or `high`). When the `BaseType` is a string we must use the `oneAlloc` function because reading a string requires allocation.

# Conclusion

Zig is my first real experience using powerful meta-programming and the experience has been mostly positive. Despite the general lack of documentation around Zig I didn’t have much trouble learning about `comptime` by looking at the existing source code in the standard library (specifically the [`std.fmt.format`](https://github.com/ziglang/zig/blob/b08d32ceb5aac5b1ba73c84449c6afee630710bb/lib/std/fmt.zig#L73) function and the [JSON parser](https://github.com/ziglang/zig/blob/b08d32ceb5aac5b1ba73c84449c6afee630710bb/lib/std/json.zig#L1917-L1926)). If you’re interested in `comptime` I encourage you to look at the standard library too.

I started this project in October 2020, it wasn’t always fun (I stumbled on quite a few compiler bugs) but I’m really happy with what I ended up with.

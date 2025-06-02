---
title: Zig, shared libraries and Musl
description:
    Article talking about how shared libraries work on Linux
date: "2025 January 18"
format: blog_entry
---

I'm currently working on upgrading the build script for [zig-sqlite](https://github.com/vrischmann/zig-sqlite) and learned something about shared libraries and musl.

For context, zig-sqlite provides a Zig wrapper for SQLite which facilitates using SQLite in Zig.

One of the things you can do with it is building a SQLite [loadable extension](https://www.sqlite.org/loadext.html) (I've written an entire [blog post](/blog/virtual-tables-with-zig-sqlite) about this).
These extensions are shared libraries that can be loaded in various ways by SQLite and as you can imagine the shared library is operating system and architecture dependant (it uses `dlopen` and equivalent under the hood).

Well, one thing I didn't realize until today was that shared libraries built with musl are _not_ loadable by _glibc_ and vice-versa (in most cases; see the [musl FAQ](https://www.musl-libc.org/faq.html)).

This incompatibility hit me today because I've been testing my changes with the `native-linux` target which defaults to using musl instead of glibc, and while the build is successful the shared library doesn't work with glibc.
You can see this with `ldd`:
```bash
$ ldd zig-out/lib/libzigcrypto.so
zig-out/lib/libzigcrypto.so: error while loading shared libraries: /lib64/libc.so: invalid ELF header
```

In my case, I saw this error too when trying to load the extension with `sqlite3`:
```bash
$ sqlite3
SQLite version 3.46.1 2024-08-13 09:16:08
sqlite> .load ./zig-out/lib/libzigcrypto
Error: /lib64/libc.so: invalid ELF header
```

I had no idea what this meant at first but [this comment](https://github.com/ziglang/zig/issues/16624#issuecomment-2294811562) on the Zig repository is where I learned that:
* there are, in fact, two different dynamic linkers for `musl` and `glibc`
* `ldd` only works with `glibc`-based binaries or libraries

Fortunately, solving this is trivial: I just had to build with the `native-linux-gnu` target.

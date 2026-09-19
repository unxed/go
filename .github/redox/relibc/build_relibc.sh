#!/bin/sh
# Runs INSIDE the redoxos/redoxer container with a relibc checkout mounted at
# /src. Builds relibc's libs (libc.a, libc.so which also is the dynamic linker
# ld64.so.1, crt objects) for x86_64-unknown-redox.
#
# Attempt 1 uses redoxer's own toolchain (`redoxer env` puts its cargo/rustc,
# the cross gcc/ar/ld and TARGET in place). relibc needs -Zbuild-std, which
# needs the rust-src component; if the redoxer toolchain lacks it, attempt 2
# uses relibc's pinned nightly + rust-src from rustup with the same cross tools.
set -x
cd /src
export TARGET=x86_64-unknown-redox
date
# `make libs` also generates the C headers with cbindgen (not in the image)
cargo install cbindgen --locked --root /opt/cbindgen > /src/cbindgen-install.log 2>&1 || { tail -30 /src/cbindgen-install.log; exit 1; }
export PATH=/opt/cbindgen/bin:$PATH
cbindgen --version
date
redoxer env sh -c 'echo PATH=$PATH; which x86_64-unknown-redox-gcc; rustc --version; cargo --version; ls "$(rustc --print sysroot)/lib/rustlib" "$(rustc --print sysroot)/lib/rustlib/src" 2>&1 | head; nproc; free -m'
if redoxer env make -j2 PROFILE=release libs > /src/build1.log 2>&1; then
  echo "=== relibc built with the redoxer toolchain"
else
  echo "=== attempt 1 (redoxer toolchain) failed; errors and tail of log:"; grep -n -i "error\|\*\*\*\|undefined\|cannot" /src/build1.log | grep -v "^.*WARN" | tail -25; tail -15 /src/build1.log
  date
  rustup toolchain install nightly-2026-05-24 --profile minimal -c rust-src
  if redoxer env make -j2 PROFILE=release CARGO="rustup run nightly-2026-05-24 cargo" libs > /src/build2.log 2>&1; then
    echo "=== relibc built with nightly-2026-05-24"
  else
    echo "=== attempt 2 failed; tail of log:"; tail -60 /src/build2.log
    exit 1
  fi
fi
date
ls -la /src/target/x86_64-unknown-redox/release/

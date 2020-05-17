# Why Amino?

* better mapping to OOP languages than proto2/3; aka "proto3/4 Any wants to be
  Amino" --> now is the time to try different approaches.  we already have a
way that work, and the usage of codecs to register types at the app level is
ultimately a necessary interface.  see
https://developers.google.com/protocol-buffers/docs/proto3#any
"https://developers.google.com/protocol-buffers/docs/proto3#any".
* go-amino specifically written such that code serves as spec (better for
  immutable code), and any determinism enforced, etc.
* faster prototype -> production cycle, future compat w/ proto3 fields.
  (not fully supported yet in Amino).

# TODOs

* `genproto/*` to generate complementary proto schema files (for support in other languages)
* use `genproto/*` generated tooling to encode/decode.
  * [ ] use both fuzz tests to check for completeness.
  * [ ] automate the testing of gofuzz tests.

# NOTES

* Code generation convention is OK here:
  `https://github.com/golang/protobuf/blob/master/protoc-gen-go/generator/generator.go`,
but shouldn't there be a better way?  Perhaps one that uses the AST, so that
the template can be checked by the compiler, even.

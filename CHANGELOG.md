# Changelog

## 0.9.7 (April 25, 2019)

FEATURES:
 - Add MustUnmarshalBinary and MustUnmarshalBinaryBare to the Codec
   - both methods are analogous to their marshalling counterparts
   - both methods will panic in case of an error
 - MarshalJSONIndent

## 0.9.6 (April 5, 2018)

IMPROVEMENTS:
 - map[string]<any> support for Amino:JSON

## 0.9.5 (April 5, 2018)

BREAKING CHANGE:
 - Skip encoding of "void" (nil/empty) struct fields and list elements, esp empty strings

IMPROVEMENTS:
 - Better error message with empty inputs

## 0.9.4 (April 3, 2018)

BREAKING CHANGE:
- Treat empty slices and nil the same in binary

IMPROVEMENTS:
- Add indenting to aminoscan

BUG FIXES:
- JSON omitempty fix.

## 0.9.2 (Mar 24, 2018)

BUG FIXES:
 - Fix UnmarshalBinaryReader consuming too much from bufio.
 - Fix UnmarshalBinaryReader obeying limit.

## 0.9.1 (Mar 24, 2018)

BUG FIXES:
 - Fix UnmarshalBinaryReader returned n

## 0.9.0 (Mar 10, 2018)

BREAKING CHANGE:
 - wire -> amino
 - Protobuf-like encoding
 - MarshalAmino/UnmarshalAmino

## 0.8.0 (Jan 28, 2018)

BREAKING CHANGE:
 - New Disamb/Prefix system
 - Marshal/Unmarshal Binary/JSON
 - JSON is a shim but PR incoming

## 0.7.2 (Dec 5, 2017)

IMPROVEMENTS:
 - data: expose Marshal and Unmarshal methods on `Bytes` to support protobuf
 - nowriter: start adding new interfaces for improved technical language and organization

BUG FIXES:
 - fix incorrect byte write count for integers

## 0.7.1 (Oct 27, 2017)

BUG FIXES:
 - dont use nil for empty byte array (undoes fix from 0.7.0 pending further analysis)

## 0.7.0 (Oct 26, 2017)

BREAKING CHANGE:
 - time: panic on encode, error on decode for times before 1970
 - rm codec.go

IMPROVEMENTS:
 - various additional comments, guards, and checks

BUG FIXES:
 - fix default encoding of time and bytes
 - don't panic on ReadTime
 - limit the amount of memory that can be allocated

## 0.6.2 (May 18, 2017)

FEATURES:

- `github.com/tendermint/go-data` -> `github.com/tendermint/go-wire/data`

IMPROVEMENTS:

- Update imports for new `tmlibs` repository

## 0.6.1 (April 18, 2017)

FEATURES:

- Size functions: ByteSliceSize, UvarintSize
- CLI tool 
- Expression DSL
- New functions for bools: ReadBool, WriteBool, GetBool, PutBool
- ReadJSONBytes function


IMPROVEMENTS:

- Makefile
- Use arrays instead of slices
- More testing
- Allow omitempty to work on non-comparable types

BUG FIXES:

- Allow time parsing for seconds, milliseconds, and microseconds
- Stop overflows in ReadBinaryBytes


## 0.6.0 (January 18, 2016)

BREAKING CHANGES:

FEATURES:

IMPROVEMENTS:

BUG FIXES:


## Prehistory


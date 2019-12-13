# Capsule Type Operation Definitions

As described in [the main introduction to Capsule types](./types.md#capsule-types),
by default a capsule type supports no operations apart from comparison by
reference identity.

However, there are certain operations that calling applications reasonably
expect to be able to use generically across values of any type, e.g. as part of
decoding arbitrary user input and preparing it for use.

To support this, calling applications can optionally implement some additional
operations for their capsule types. This is limited only to the subset of
operations that, as noted above, are reasonable to expect to have available on
values of any type.

It does not include type-specialized operations like
arithmetic; to perform such operations against capsule types, implement that
logic in your calling application instead where you must presumably already
be making specialized decisions based on value types.

The following operations are implementable:

* The `GoString` result for values of the type, so that values can be included
  in `fmt` operations using `%#v` along with other values and give a more
  useful result.

  To stay within `cty`'s conventions, the `GoString` result should generally
  represent a call to a value constructor function that would produce an
  equivalent value.

* The `GoString` result for the type itself, for similar reasons.

* Equality as an operation. It's unnecessary to implement this unless your
  capsule type represents a container for other `cty` values that would
  need to be recursively compared for equality. For simple capsule types,
  just implement raw equality and it will be used for both situations.

* Raw equality as an integration method. This is commonly used in tests in
  order to take into account not only the value itself but its null-ness and
  unknown-ness. Because capsule types are primarily intended to be simple
  transports for opaque application values, in simple cases this integration
  method can just be a wrapper around whatever normal equality operation would
  apply to the wrapped type.

* Conversion to and from the capsule type, using the `convert` package. Some
  applications use conversion as part of decoding user input in order to
  coerce user values into an expected type, in which case implementing
  conversions can make an application's capsule types participate in such
  coersions as needed.

To implement one or more of these operations on a capsule type, construct it
with `cty.CapsuleWithOps` instead of just `cty.Capsule`.

The operation implementations are provided as function references within a
struct value, and so those function references can potentially be closures
referring to arguments passed to a type constructor function in order to
implement parameterized capsule types.

For more information on the available operations and the contract for
implementing each one, see the documentation on the fields of `cty.CapsuleOps`.

## Extension Data

In addition to the operations that `cty` packages use directly as described
above, `CapsuleOps` includes extra function `ExtensionData` which can be used
for application-specific extensions.

It creates an extensible namespace using arbitrary comparable values of named
types, where an application can define a key in its own package namespace and
then use it to retrieve values from cooperating capsule types.

For example, if there were an application package called `happyapp` which
defined an extension key `Frob`, a cooperating capsule type might implement
`ExtensionData` like this:

```go
    ExtensionData: func (key interface{}) interface{} {
        switch key {
        case happyapp.Frob: // a value of some type belonging to "happyapp"
            return "frob indeed"
        default:
            return nil
        }
    }
```

Package `happyapp` should define this `Frob` symbol as being of a named type
belonging to its own package, so that the type can serve as a namespace:

```go
type frobType int

var Frob frobType
```

`cty` itself doesn't do anything special with `ExtensionData`, but there is a
convenience method `CapsuleExtensionData` on `cty.Type` that can be used with
capsule types to attempt to access extension data keys without first retrieving
the `CapsuleOps` and checking if it implements `ExtensionData`:

```go
// (in package happyapp)
if frobValue, ok := givenType.CapsuleExtensionData(Frob).(string); ok {
    // If we get in here then the capsule type correctly handles Frob,
    // returning a string. If the capsule type does not handle Frob or
    // if it returns the wrong type then we'll ignore it.
    log.Printf("This capsule type has a Frob: %s", frobValue)
}
```

`CapsuleExtensionData` will panic if called on a non-capsule type.

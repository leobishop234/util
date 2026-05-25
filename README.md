# Batch 

This package provides helpers for safe, predictable batch operations on maps and slices.

## naming convention

- `Get...`: read a value from an existing map by key.
- `Transform...`: apply a value transformation while preserving collection shape.
- `Index...`: build a map from a slice using a key selector.
- `Group...`: build a map of slices from a slice using a key selector.
- `...By...`: indicates the function uses a key selector.
- `...To...`: indicates the function both uses a key selector and transforms the values.

E.g.
- `IndexBy...`: build `map[K]V` from `[]V` using a key selector.
- `IndexTo...`: build `map[K]O` from `[]I` with both key selection and value transformation.
- `GroupBy...`: build `map[K][]V` from `[]V` by key.
- `GroupTo...`: build `map[K][]O` from `[]I` by key with value projection.

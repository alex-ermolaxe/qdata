# qdata MCP usage

- Start every workflow by calling `load` and retain the returned `session_id`.
- Use a distinct session for each file when working with multiple datasets.
- Inspect available fields with `schema` before writing filters or selecting fields.
- After applying a filter with `where`, call `count` to verify the result size.
- `where` syntax is `<field> <operator> <value>`; supported operators are `=`, `!=`, `>`, `<`, `>=`, `<=`, `~`, `!~`, `^`, `$`, `IN`, and `EXISTS`. Join conditions with `AND` or `OR`; address nested fields with dot paths such as `address.city`.
- Pass the `session_id` to every tool that operates on a loaded dataset.

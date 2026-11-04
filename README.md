# qdata

**qdata** - an interactive command-line tool for querying, filtering and transforming
structured data files (JSON, XML, CSV) without writing any code. Load a file,
filter records, pick or exclude fields, sort results and save the output —
all in a simple SQL-like syntax.

## Table of Contents

- [WHERE — Filter records](#where--filter-records)
- [SELECT — Pick fields](#select--pick-fields)
- [EXCLUDE — Remove fields](#exclude--remove-fields)
- [SORT — Sort records](#sort--sort-records)
- [SHOW — Display results](#show--display-results)
- [SAVE — Save to file](#save--save-to-file)
- [FLATTEN — Flatten nested object fields](#flatten--flatten-nested-object-fields)
- [TRANSFORM — Transform records to maps](#transform--transform-records-to-maps)
- [RESET — Reset to original](#reset--reset-to-original)
- [COUNT — Count records](#count--count-records)
- [SCHEMA — View data schema](#schema--view-data-schema)
- [INFO — List commands](#info--list-commands)
- [EXIT — Quit session](#exit--quit-session)
- [VERSION — Show application version](#version--show-application-version)
- [Using with AI Assistants (MCP)](#using-with-ai-assistants-mcp)

---

## Getting Started

Launch the application by pointing it to a file:

```bash
qdata --file ./data.json

# Explicit format
qdata --file ./data.json --format json
```

After launch, an interactive session opens:

```
qdata v1.0 | file: data.json | records: 1500
>
```

Every command operates on the **current intermediate result**. The result is updated after each command and is available for further querying. Use [`RESET`](#reset--reset-to-original) to return to the original dataset at any time.

---

## Using with AI Assistants (MCP)

qdata connects to MCP clients such as GitHub Copilot, Cursor, Claude Desktop, and others by running `qdata mcp-serve`. It provides tools for loading, filtering, sorting, and transforming data, and supports working with multiple files at once through named, isolated sessions.

### Installation

Install qdata into your Go binary directory:

```bash
go install github.com/alex-ermolaxe/qdata@latest
```

Make sure `$(go env GOPATH)/bin` is included in your `PATH`. Alternatively, build the binary locally and use its absolute path in your MCP client configuration:

```bash
go build -o qdata .
```

### Configuration

#### GitHub Copilot (JetBrains / VS Code)

```json
{
  "servers": {
    "qdata": {
      "command": "qdata",
      "args": ["mcp-serve"]
    }
  }
}
```

#### Cursor (`~/.cursor/mcp.json`)

```json
{
  "mcpServers": {
    "qdata": {
      "command": "qdata",
      "args": ["mcp-serve"]
    }
  }
}
```

#### Claude Desktop (`claude_desktop_config.json`)

```json
{
  "mcpServers": {
    "qdata": {
      "command": "qdata",
      "args": ["mcp-serve"]
    }
  }
}
```

If qdata is not in `PATH`, replace `"command": "qdata"` with the absolute path to the executable. Keep `"args": ["mcp-serve"]` in every configuration so the MCP server starts instead of the interactive CLI.

### Available MCP Tools

| Tool | Purpose |
|------|---------|
| `load` | Load a data file into a named session. |
| `list_sessions` | List loaded sessions, source paths, and current record counts. |
| `close_session` | Close a session and discard its current query state. |
| `where` | Filter records by a condition. |
| `select` | Keep specified fields. |
| `exclude` | Remove specified fields. |
| `sort` | Sort records by a field. |
| `show` | Display current records as JSON. |
| `save` | Save current records beside the source file. |
| `flatten` | Flatten a nested object field. |
| `transform` | Transform records into maps using selected key and value paths. |
| `reset` | Restore the session's original records. |
| `count` | Count records in the current result. |
| `schema` | Show available fields and inferred types. |

All tools except `load` and `list_sessions` require `session_id`, which is returned by `load`.

### Usage notes

- Start with `load`; it returns the `session_id` to use in subsequent calls.
- Use a different `session_id` for each file; sessions are isolated from one another.
- Provide absolute paths to data files.
- MCP mode does not affect the regular `qdata --file` interactive CLI.

### Troubleshooting

| Symptom | Solution |
|---------|----------|
| MCP tools do not appear | Check the executable path and that the configuration contains `"args": ["mcp-serve"]`, then restart the MCP client. |
| Server exits immediately | Make sure it is launched with `mcp-serve`, not as plain `qdata`. |
| A data file does not load | Use the file's absolute path. |

---

## WHERE — Filter records

Filters the current record set by a condition. The result becomes the new intermediate result.

```
WHERE <condition> [AND|OR <condition> ...]
```

### Operators

| Operator | Description |
|----------|-------------|
| `=`      | Strict equality |
| `!=`     | Not equal |
| `>`      | Greater than |
| `<`      | Less than |
| `>=`     | Greater than or equal |
| `<=`     | Less than or equal |
| `~`      | Contains substring (case-insensitive) |
| `!~`     | Does not contain substring |
| `^`      | Starts with |
| `$`      | Ends with |
| `IN`     | Value is in list |
| `EXISTS` | Field exists in record |

### Examples

```
-- Simple equality
WHERE status = "active"

-- Numeric comparison
WHERE age > 30

-- Substring match (case-insensitive)
WHERE name ~ "john"

-- Nested field
WHERE address.city = "Moscow"

-- Logical AND
WHERE age > 25 AND status = "active"

-- Logical OR
WHERE department = "IT" OR department = "HR"

-- Value in list
WHERE status IN ["active", "pending"]

-- Field existence check
WHERE phone EXISTS

-- Grouped conditions
WHERE (age > 18 AND age < 65) AND status = "active"
```

### Output

```
> WHERE age > 30 AND status = "active"
✓ Found: 243 records (was: 1500)
```

---

## SELECT — Pick fields

Keeps only the specified fields in each record.

```
SELECT <field1>, <field2>, ...
```

### Examples

```
-- Pick specific fields
SELECT id, name, email

-- Pick nested fields
SELECT id, name, address.city, address.zip

-- Pick fields from each object in an array
SELECT id, items.name, items.price

-- Reset to full record
SELECT *
```

### Notes

- Nested fields such as `address.city` are extracted and placed into the result.
- For arrays of objects, paths such as `items.name` keep the array structure and include only the selected fields in each object.

### Output

```
> SELECT id, name, address.city
✓ Applied SELECT to 243 records
```


Keeps only the specified fields in each record.

```
SELECT <field1>, <field2>, ...
```

### Examples

```
-- Pick specific fields
SELECT id, name, email

-- Pick nested fields
SELECT id, name, address.city, address.zip

-- Reset to full record
SELECT *
```

### Notes

- Nested fields such as `address.city` are extracted and placed into the result.
- The original nested structure is preserved in the output.

### Output

```
> SELECT id, name, address.city
✓ Applied SELECT to 243 records
```

---

## EXCLUDE — Remove fields

Removes the specified fields from each record.

```
EXCLUDE <field1>, <field2>, ...
```

### Examples

```
-- Exclude sensitive fields
EXCLUDE password, token, internal_id

-- Exclude nested fields
EXCLUDE user.password, user.salt

-- Exclude an entire nested object
EXCLUDE metadata
```

### Output

```
> EXCLUDE password, created_at
✓ Applied EXCLUDE to 243 records
```

---

## SORT — Sort records

Sorts the current record set by the specified field.

```
SORT <field> [ASC|DESC]
```

- Default direction is `ASC` when omitted.
- Numbers are sorted numerically, strings are sorted lexicographically.
- Records missing the specified field are placed at the end.

### Examples

```
-- Ascending (default)
SORT age

-- Descending
SORT created_at DESC

-- Nested field
SORT address.city ASC

-- String field
SORT name ASC
```

### Output

```
> SORT age DESC
✓ Sorted 243 records by "age" DESC
```

---

## SHOW — Display results

Prints the current intermediate result to the terminal.

```
SHOW [LIMIT <n>] [OFFSET <n>]
```

### Examples

```
-- Show all records
SHOW

-- Show first 10 records
SHOW LIMIT 10

-- Show 10 records starting from position 20
SHOW LIMIT 10 OFFSET 20
```

### Output

```
> SHOW LIMIT 2
[
  { "id": 1, "name": "Alice", "age": 32 },
  { "id": 2, "name": "Bob",   "age": 28 }
]
── 2 of 243 records ──
```

---

## SAVE — Save to file

Saves the current intermediate result to a file in the same directory as the source file.

```
SAVE [AS <filename>] [FORMAT <format>]
```

- When `AS` is omitted, the file is saved as `<original_name>_result.<ext>`.
- When `FORMAT` is omitted, the original file format is used.
- If the target file already exists, a confirmation prompt is shown before overwriting.

### Examples

```
-- Auto-generated filename (e.g. data_result.json)
SAVE

-- Custom filename
SAVE AS filtered_users

-- Custom filename and format
SAVE AS report FORMAT csv
```

### Output

```
> SAVE AS active_users
✓ Saved 243 records → /path/to/active_users.json
```

---

## FLATTEN — Flatten nested object fields

FLATTEN expands all fields of a nested object into the top-level record. If any field names collide with existing top-level fields, all flattened fields will be prefixed with the object field name followed by an underscore.

```
FLATTEN <object_field>
```

### Examples

```
-- Simple flatten
FLATTEN address

-- Flatten multiple nested objects
FLATTEN address
FLATTEN work
```

### Behavior

- If the top-level object already contains a field with the same name as a nested field, the flattened fields are added with the prefix `<object_field>_` (e.g. `work_city`).
- Non-object fields cause an error (e.g. `FLATTEN name` when `name` is a string).
- If the specified object field does not exist, the dataset is left unchanged.

### Output

```
> FLATTEN address
✓ Flattened field "address" in 1500 records (was: 1500)
```

---

## TRANSFORM — Transform records to maps

Creates one output record per current record. The value in the `KEY` field becomes the output object's property name, and the `VALUE` field supplies its value. Paths through arrays of objects return an array of the matching values.

```
TRANSFORM TO-MAP KEY <key_field> VALUE <value_field>
```

For example, given a record with `sap_order_id` and an `assets` array:

```
TRANSFORM TO-MAP KEY sap_order_id VALUE assets.name
```

produces a record equivalent to:

```json
{
  "000001810600": ["K06", "K610"]
}
```

Each transform replaces the current records with the transformed records. Both paths must exist; the key value must be a non-null scalar. To map an entire array, use its field as the value path, for example `VALUE assets`.

---

## RESET — Reset to original

Discards all applied filters and transformations, restoring the full original dataset.

```
RESET
```

### Output

```
> RESET
✓ Reset to original: 1500 records
```

---

## COUNT — Count records

Prints the number of records in the current intermediate result.

```
COUNT
```

### Output

```
> COUNT
243 records
```

---

## SCHEMA — View data schema

Displays all available fields and their inferred types. The schema is analysed across the first N records of the current dataset.

```
SCHEMA
```

### Output

```
> SCHEMA
Fields in current dataset:
  id              number
  name            string
  email           string
  age             number
  status          string
  address
    └─ city       string
    └─ zip        string
    └─ country    string
  tags            array<string>
  created_at      string
```

---

## INFO — List commands

Prints the available interactive commands and the arguments each command accepts.

```
INFO
```

---

## EXIT — Quit session

Closes the interactive session.

```
EXIT
```

---

## VERSION — Show application version

The application exposes a standard Cobra version command and flag. Use any of the following to print the current qdata version and exit:

```
qdata --version
qdata version
qdata -v
```

Example output:

```
qdata v1.1.1
```

---

## Autocomplete

Press `Tab` at any point to trigger autocomplete.

| Context | Suggestions |
|---------|-------------|
| Beginning of input | Command keywords: `WHERE`, `SELECT`, `EXCLUDE`, `SORT`, `SHOW`, `SAVE`, `FLATTEN`, `TRANSFORM`, `RESET`, `COUNT`, `SCHEMA`, `INFO`, `EXIT` |
| After a command keyword | Field names from the current dataset schema |
| After a dot in a field path | Nested field names (e.g. `address.` → `address.city`, `address.zip`) |
| After `SORT <field>` | `ASC`, `DESC` |
| After `SAVE FORMAT` | Registered format names: `json`, `xml`, `csv` |
| After `IN` | Opening bracket `[` |

```
> WHERE add[TAB]
address.city    address.zip    address.country
```

---

## Full Session Example

```
$ qdata --file ./users.json
qdata v1.0 | file: users.json | records: 1500

> WHERE status = "active" AND age > 25
✓ Found: 487 records (was: 1500)

> WHERE address.city = "Moscow"
✓ Found: 102 records (was: 487)

> SORT age DESC
✓ Sorted 102 records by "age" DESC

> EXCLUDE password, internal_token, metadata
✓ Applied EXCLUDE to 102 records

> SELECT id, name, email, age, address.city
✓ Applied SELECT to 102 records

> SHOW LIMIT 3
[
  { "id": 42,  "name": "Alice",   "email": "alice@example.com", "age": 61, "address.city": "Moscow" },
  { "id": 105, "name": "Mikhail", "email": "misha@example.com", "age": 58, "address.city": "Moscow" },
  { "id": 78,  "name": "Elena",   "email": "elena@example.com", "age": 55, "address.city": "Moscow" }
]
── 3 of 102 records ──

> SAVE AS moscow_active_users
✓ Saved 102 records → /data/moscow_active_users.json

> EXIT
Bye!
```
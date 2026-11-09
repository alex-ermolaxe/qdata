# Code readability

- For new and modified code, prefer named intermediate variables when they make nested calls or multi-step return expressions easier to read and debug.
- Keep error handling explicit; do not introduce trivial helper functions solely to shorten a return expression.
- In Go, put a blank line immediately before `return` only when another statement precedes it in the same block (function body, conditional or loop body, switch or select case). Do not add a blank line when `return` is the block's first statement.
- When a struct literal has more than one field, put each field on its own line.

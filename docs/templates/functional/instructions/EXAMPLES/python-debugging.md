name: python-debugging-instructions
version: 1.0.0
status: stable
applies_to:

- "\*_/_.py"
  deprecation: null

<!-- markdownlint-disable-file -->

# Instructions: Python Debugging (VS Code)

## When Invoked

Automatically loaded when editing `.py` files in VS Code.

## Guidance

### DO's ✅

- Use `import pdb; pdb.set_trace()` for breakpoints
- Run with `python -m pdb script.py` for line-by-line debugging
- Check stack trace for error location first

### DON'Ts ❌

- Don't leave `pdb` statements in production code
- Don't rely only on print() for debugging
- Don't commit debug code

### Tools Available

- **Pylance**: Type checking, quick fixes
- **Python Debugger**: Built-in VS Code
- **pytest**: Run tests with `-vv` for details

### Links

- [Python pdb docs](https://docs.python.org/3/library/pdb.html)
- [VS Code Python guide](https://code.visualstudio.com/docs/python/debugging)

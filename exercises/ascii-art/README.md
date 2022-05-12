
This program reads a "shifted" ASCII art from the `ascii-art.txt` file
alongside a mapping contained in the file `mapping.txt`.

A "shifted" ASCII art is equal to the original one, except for the position
of the lines: they have been moved following the mapping contained in the other
file.

The mapping file contains a list of key-value pairs with the following format:

```
original-line-position -> current-line-position
```

The key (*original-line-position*) is the original position of the line in the
ASCII art, while the value (*current-line-position*) is the current position of that line in the file.

Therefore, the first line of the file:

```
0: 6
```

means that the first line in the original ASCII art is now the sixth in the
"shifted" ASCII art.

The program should be able to reconstruct the original ASCII art, but the
output seems meaningless. Can you find out why?

<details>
  <summary>Hint</summary>

  Look at how the rows from the "shifted" ASCII art are loaded from the file.
  Try to inspect what the function is doing and how it is building the output slice.
</details>

<details>
  <summary>Solution</summary>

  The `rows` slice declared in line 71 has been incorrectly created with a backing array of 27 strings. At line 75 we are using the `append` built-in, so the first 27 lines of the returned slice, used in the calling function, are empty.

  Substitute:

  ```go
  rows := make([]string, 27)
  ```

  with:

  ```go
  var rows []string
  ```
</details>

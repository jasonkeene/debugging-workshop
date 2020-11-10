
This program generates random strings and then counts how often the first six
characters of the strings are seen. Since there are only three letters used
in the random strings there should only be a total of 3^6 prefixes. We indeed
see it only generates 729 different prefixes when it runs but it consumes well
over a GiB of memory! 729 unique six-character strings and their counters
should total a few KiB. Why is this taking up so much memory?

<details>
  <summary>Hint</summary>
  
  Use pprof to take a memory profile. What lines of code are using the most
  memory? How can they be reduced?
</details>

<details>
  <summary>Solution</summary>
  
  The original, large rune slice is garbage collected but a copy is made when
  it is converted to a string. We only use the first few characters of the
  string when we slice it but the new string points to the same data as the
  original string and so it can not be garbage collected.
</details>

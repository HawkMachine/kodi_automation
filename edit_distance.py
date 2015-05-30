"""Module provides a function to compute edit distance between strings."""

import argparse


def EditDistance(first, second):
  """Returns edit distance between given strings.

  Edit distance is defined as minimal number of edits that transforms first
  string into the second one.

  Possible operations:
    - Remove character.
    - Add character.
    - Substitute character.

  Args:
    first: str
    second: str

  Returns:
    int, edit distance between the first and second string.
  """

  if first is None or second is None:
    raise ValueError('EditDistance argument: cannot be None')

  # To compute the edit distance between two strings the dynamic algorithm is
  # used.
  # Edits can be sorted from left to right. Each letter is touched only once:
  # added, removed, or changed. Let's say that we first remove, then add, then
  # then change letters and we do it from left to right.

  # distance table:
  #  Idea: distance[i][j] = edit distance between first[:i] and second[:j]
  #  distance[:len(first)][:len(second)] - edit distance between first and second.
  # Recursive definition:
  #  distance[0][x] = x
  #  distance[x][0] = x
  #  distance[i][j] (i > 0, j > 0)
  #  = minimum between:
  #    * (remove character) distance[i-1][j] + 1
  #    * (add character) distance[i][j-1] + 1
  #    * (substitute character) distance[i-1][j-1] + 1
  #    * (only if last characters are the same) distance[i-1][j-1]
  # Definition based entirely on the row or column if filled before.
  distance = []
  for unused_c in xrange(len(first) + 1):
    distance.append([0] * (len(second) + 1))

  # Fill distance[0][x]
  for j in xrange(len(second) + 1):
    distance[0][j] = j

  # Fill distance[x][0]
  for i in xrange(len(first) + 1):
    distance[i][0] = i

  for i in xrange(1, len(first) + 1):
    for j in xrange(1, len(second) + 1):
      distance[i][j] = min(
          distance[i-1][j] + 1,
          distance[i][j-1] + 1,
          distance[i-1][j-1] + 1)
      if first[i-1] == second[j-1]:
        distance[i][j] = min(
            distance[i-1][j-1],
            distance[i][j])

  return distance[len(first)][len(second)]


def main():
  parser = argparse.ArgumentParser()
  parser.add_argument('first')
  parser.add_argument('second')
  args = parser.parse_args()
  dist = EditDistance(args.first, args.second)
  print dist


if __name__ == '__main__':
  main()

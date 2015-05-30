#!/usr/bin/python
"""Module contains a couple of classes that try to classfy
filenames to tv-shows.
"""

import os
import argparse
import re

from kodi_automation import episode_classification
from kodi_automation import edit_distance

def Normalize(show_name):
  """Returns a normalized show name."""
  if show_name is None:
    raise ValueError('Argument cannot be None')

  # First replace unwanted characters into spaces.
  show_name = re.sub(r'\.', ' ', show_name)
  show_name = re.sub(r'_', ' ', show_name)

  # Now normilize spaces.
  show_name = re.sub(r'\s+', ' ', show_name)
  show_name = re.sub(r'^\s+', '', show_name)
  show_name = re.sub(r'\s+$', '', show_name)

  # Make the show look nice.
  show_name = show_name.title()
  return show_name


class ShowNameClassificator(object):
  """Filename classificator."""

  SINGLE_EPISODE_REGEXPS = [
      # Season and episode explicitly in the filename
      re.compile(r'(?P<show_name>.*)[- _\.]+season[- _]+(?P<season_nr>\d+)[- _x]episode[- _]+(?P<episode_nr>\d+).*',
                 re.IGNORECASE),
      # S and E present
      re.compile(r'(?P<show_name>.*)[- _\.]+s(?P<season_nr>\d+)[_xX ]?e(?P<episode_nr>\d+).*',
                 re.IGNORECASE),
      # S and E missing, must be some kind of delimeter
      re.compile(r'(?P<show_name>.*)[- _\.]+(?P<season_nr>\d+)[_xX ](?P<episode_nr>\d+).*',
                 re.IGNORECASE),
  ]


  MISSING_SEASON_SINGLE_EPISODE_REGEXPS = [
      # Assume season 2 when season is not present
      re.compile(r'(?P<show_name>.*)[ _-]+(?:e|ep|ep\.|episode)[ _-](?P<episode_nr>\d+).*',
                 re.IGNORECASE),
  ]

  def Classify(self, path):
    """Classifies given path by using a set of predefined refexpes."""
    if not path:
      raise ValueError('Argument cannot be None')

    path = os.path.split(path)[1]

    show_name = None
    season_nr = None
    episode_nr = None
    found = False
    for regexp in self.SINGLE_EPISODE_REGEXPS:
      match = regexp.match(path)
      if match:
        found = True
        show_name = match.group('show_name')
        season_nr = match.group('season_nr')
        episode_nr = match.group('episode_nr')
        break
    if found:
      return show_name, int(season_nr), int(episode_nr)

    for regexp in self.MISSING_SEASON_SINGLE_EPISODE_REGEXPS:
      match = regexp.match(path)
      if match:
        found = True
        show_name = match.group('show_name')
        season_nr = '1'
        episode_nr = match.group('episode_nr')
    if found:
      return show_name, int(season_nr), int(episode_nr)

    return None, None, None


class EditDistanceClassificator(object):

  def __init__(self, paths_with_show):
    self.paths_with_show = paths_with_show

  def Classify(self, path):
    if not path:
      raise ValueError('Argument cannot be None')

    path = os.path.split(path)[1]

    # edit distance should be smaller than this.
    max_data_length = max(len(d) for d in self.paths_with_show)
    current_distance = len(path) + max_data_length + 1
    current_show_name = None

    for filename, show_name in self.paths_with_show.iteritems():
      distance = edit_distance.EditDistance(path, filename)
      if distance < current_distance:
        current_distance = distance
        current_show_name = show_name
    return current_show_name


def main():
  parser = argparse.ArgumentParser()
  parser.add_argument('paths', nargs='+')

  args = parser.parse_args()

  for path in args.paths:
    try:
      show, season, episode = Classification(path)
    except:
      pass
    print season, episode, show


if __name__ == '__main__':
  main()

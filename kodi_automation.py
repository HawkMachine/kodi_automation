"""Automation of moving files to kodi-scanned directories."""

import os
import sys
import argparse
import Queue


def ScanDir(dirpath, exts=None):
  """Scans direcotry and returns files."""
  # exts == None - return all

  queue = Queue.Queue()
  queue.put(dirpath)

  result = []

  while not queue.empty():
    dirpath = queue.get()
    for _, dirnames, filenames in os.walk(dirpath):
      for new_dirpath in dirnames:
        queue.put(os.path.join(dirpath, new_dirpath))

      # Filters paths if required.
      if exts:
        filenames = [filepath for filepath in filenames
                     if any(filepath.endswith(ext) for ext in exts)]
      result.extend(filenames)

  return result


def Classification(paths):
  return ([], [])


def MoveMoveFile(path, movies_dir, dry_run=False):
  if dry_run:
    sys.stderr.write('Moving movie', path)
    return


def MoveEpisodeFile(path, seria, season, episode, series_dir, dry_run=False):
  if dry_run:
    sys.stderr.write('Moving episode', *args)
    return


def main():
  parser = argparse.ArgumentParser()
  parser.add_argument('--scan-dir', '-s', dest='scan_dir', default=None)
  parser.add_argument('--movies-dir', dest='movies_dir', default=None)
  parser.add_argument('--series-dir', dest='series_dir', default=None)
  parser.add_argument('--video-exts', '-v', dest='video_exts',
                      default='mkv,avi,mp4')
  parser.add_argument('--dry-run', dest='dry_run', default=False)
  args = parser.parse_args()

  if not args.scan_dir or not args.movies_dir or not args.series_dir:
    parser.print_help()
    return 1

  video_exts = args.video_exts.split(',')

  new_paths = ScanDir(args.scan_dir, exts=video_exts)
  new_paths = [path for path in new_paths if any(path.endswith(ext) for ext in video_exts)]

  movies_paths, episodes = Classification(new_paths)

  for movie_path in movies_paths:
    print 'Moving', path, 'to', args.movies_dir
    MoveMoveFile(movie_path, args.movies_dir, dry_run=args.dry_run)

  for episode in episodes:
    print 'Moving', episode.path, 'as', episode.seria, 'S', episode.season, 'E', episode.episode, 'to', args.series_dir
    MoveEpisodeFile(
        episode.path, episode.seria, episode.season, episode.episode,
        args.series_dir, dry_run=args.dry_run)


if __name__ == '__main__':
  sys.exit(main())

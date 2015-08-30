import os
import sys
import sqlite3
import argparse

MISSING_MOVIES_IN_DB_CMD = 'missing_movies_in_db'
MISSING_EPISODES_IN_DB_CMD = 'missing_episodes_in_db'


def DirectoryListing(dir):
	listing = []
	for dirname, dirnames, filenames in os.walk(dir):
		for filename in filenames:
			listing.append(os.path.join(dirname, filename))
	return listing


def FilterByExtension(paths, extensions):
	return [p for p in paths if any(
		p.lower().endswith(e.lower()) for e in extensions)]

def main():
	parser = argparse.ArgumentParser()
	parser.add_argument('cmd', choices=[
		MISSING_MOVIES_IN_DB_CMD,
		MISSING_EPISODES_IN_DB_CMD],
		default=MISSING_MOVIES_IN_DB_CMD)
	parser.add_argument('--dir', required=True)
	parser.add_argument('--db',
			default=os.path.expanduser('~/.kodi/userdata/Database/MyVideos90.db'))
	parser.add_argument('--exts', dest='exts',
			default='avi,mp4,mkv')
	parser.add_argument('--format', dest='format', choices=['list', 'count'],
			default='list')

	args = parser.parse_args()

	exts = [e.lower() for e in args.exts.split(',')]

	conn = sqlite3.connect(args.db)
	cursor = conn.cursor()

	if args.cmd == MISSING_MOVIES_IN_DB_CMD:
		cursor.execute('''SELECT strPath,strFilename FROM movieview''')
		rows = cursor.fetchall()
		dbpaths = set()
		for strPath, strFilename in rows:
			if strFilename.startswith('stack://'):
				# print "***", strFilename
				strFilename = strFilename[8:]
				for p in strFilename.split(' , '):
					dbpaths.add(p.strip())
			else:	
				dbpaths.add(os.path.join(strPath, strFilename))
		dbpaths = set(p.encode('utf-8') for p in dbpaths)
	elif args.cmd == MISSING_EPISODES_IN_DB_CMD:
		cursor.execute('''SELECT strPath,strFilename FROM episodeview''')
		rows = cursor.fetchall()
		dbpaths = set()
		for strPath, strFilename in rows:
			dbpaths.add(os.path.join(strPath, strFilename))
		dbpaths = set(p.encode('utf-8') for p in dbpaths)
	else:
		return 1

	paths = set(DirectoryListing(args.dir))
	if exts:
		paths = set(FilterByExtension(paths, exts))
	missing_paths = paths - dbpaths

	if args.format == 'count':
		print len(missing_paths)
	else:
		for path in sorted(missing_paths):
			print path

if __name__ == '__main__':
	sys.exit(main() or 0)

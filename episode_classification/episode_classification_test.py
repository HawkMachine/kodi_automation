import unittest

from kodi_automation.episode_classification import episode_classification


class ShowNameClassificatorTest(unittest.TestCase):

    def setUp(self):
      self.classificator = episode_classification.ShowNameClassificator()

    def _TestClassification(self, filename, expected_show_name,
        expected_season_nr, expected_episode_nr):
      show, season_nr, episode_nr = self.classificator.Classify(filename)

      self.assertEquals(show, expected_show_name)
      self.assertEquals(season_nr, expected_season_nr)
      self.assertEquals(episode_nr, expected_episode_nr)

    def testShowNameWithSpaces(self):
      self._TestClassification(
          'Some Show Name S04E12.mkv',
          'Some Show Name', 4, 12)

    def testFilenameWithTitle(self):
      self._TestClassification(
          'Trigun/Season 1/Trigun S08E15 demons eye.avi',
          'Trigun', 8, 15)

    def testSeasonNumberMissing(self):
      self._TestClassification(
          'Berkserk ep 11.mkv',
          'Berkserk', 1, 11)

    def testOtherNumbersInFilename(self):
      self._TestClassification(
          'almost.human.s01e05.blood.brothers.720p.web.dl.sujaidr.mkv',
          'almost.human', 1, 5)

    def testXDelimeter(self):
      self._TestClassification(
          'Defying Gravity 3x11.HDTV.720p.x264.DD5.1.mkv',
          'Defying Gravity', 3, 11)

    def testSeasonAndEpisodeInFull(self):
      self._TestClassification(
          'Star Trek The Next Generation Season 5 Episode 11 - Hero Worship.avi',
          'Star Trek The Next Generation', 5, 11)

    def testUnderscores(self):
      self._TestClassification(
          'Season 1/friends_s01e11_720p_bluray_x264-sujaidr.mkv',
          'friends', 1, 11)


class EditDistanceClassificatorTest(unittest.TestCase):

  def testSimpleEditDistance(self):
    classificator = episode_classification.EditDistanceClassificator({
      'Show Name/Season 1/Show Name S01E12 title.mkv': 'Show Name',
      'Show Name/Season 2/Show Name S02E02 other stuff.mkv': 'Show Name',
      'Other Show/Season 1/Other Show S01E12.mkv': 'Other Show',
      'Other Show/Season 1/Other Show S01E12 very very very very very very looooooong title.mkv': 'Other Show',
    })

    classification_show = classificator.Classify(
      'Show Name S01E12 quality and other stuff/Show Name S02E01 new episode.mkv')
    self.assertEquals('Show Name', classification_show)


class ShowNameNormalizeTest(unittest.TestCase):

  def _TestShowNameNormalizator(self, show_name, expected_normalized_name):
    normalized_name = episode_classification.Normalize(show_name)
    self.assertEquals(normalized_name, expected_normalized_name)

  def testNormalizeSpaces(self):
    self._TestShowNameNormalizator(
        '  some    show    name    ',
        'Some Show Name')

  def testNormalizeDots(self):
    self._TestShowNameNormalizator(
        'almost.human',
        'Almost Human')

  def testNormalizeUnderscores(self):
    self._TestShowNameNormalizator(
        'almost_human',
        'Almost Human')


if __name__ == '__main__':
      unittest.main()

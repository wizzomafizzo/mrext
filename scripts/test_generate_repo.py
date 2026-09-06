"""Downloader compatibility tests; never write release databases."""

import unittest
from types import SimpleNamespace
from unittest.mock import mock_open, patch

from scripts import generate_repo


class GamesMenuReleaseTests(unittest.TestCase):
    def test_gamesmenu_binary_keeps_installed_key(self):
        with patch("builtins.open", mock_open(read_data=b"ELF fixture")), patch(
            "scripts.generate_repo.os.stat", return_value=SimpleNamespace(st_size=11)
        ):
            individual = generate_repo.create_app_db("gamesmenu", "v-test")
            combined = generate_repo.create_all_db("v-test")

        key = "Scripts/gamesmenu.sh"
        self.assertEqual(individual["db_id"], "mrext/gamesmenu")
        self.assertEqual(list(individual["files"]), [key])
        self.assertEqual(individual["files"][key], combined["files"][key])
        self.assertEqual(
            combined["files"][key]["url"],
            "https://github.com/wizzomafizzo/mrext/releases/download/v-test/gamesmenu.sh",
        )
        self.assertEqual(combined["files"][key]["tags"], ["gamesmenu"])
        self.assertFalse(combined["files"][key]["reboot"])


if __name__ == "__main__":
    unittest.main()

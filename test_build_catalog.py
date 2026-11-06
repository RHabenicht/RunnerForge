import copy
import json
import pathlib
import tempfile
import unittest

import build_catalog as catalog


class CatalogTests(unittest.TestCase):
    def test_policy(self):
        for sources, expected in [
            ([], False), (['sr3'], True), (['sr3', 'sr2'], False),
            (['tss'], False), (['r3'], False), (['unknown'], False),
        ]:
            with self.subTest(sources=sources):
                self.assertEqual(catalog.valid({'sources': sources}, catalog.APPROVED), expected)

    def test_citation_extraction(self):
        self.assertEqual(catalog.refs(['sr3.90|tss.13|newbook.12'], set()),
                         ['newbook', 'sr3', 'tss'])
        self.assertEqual(catalog.refs(['Bioware.dat target.w/SM2 1.50'], set()), [])

    def test_embedded_citation_and_stable_ids(self):
        books = [{'code': 'sr3', 'allowed': True}]
        records = [
            dict(id=8, kind='skills', sources=['sr3'],
                 fields={'specializations': ['Specialty|tss.13|']}),
            dict(id=42, kind='gear', sources=['sr3'], fields={}),
        ]
        with tempfile.TemporaryDirectory() as directory:
            output = pathlib.Path(directory) / 'catalog.json'
            excluded = pathlib.Path(directory) / 'quarantine.json'
            result = catalog.write_catalog(books, records, [], output, excluded)
            self.assertEqual([item['id'] for item in result['entries']], [42])
            self.assertEqual(json.loads(excluded.read_text())[0]['id'], 8)
            first = (output.read_bytes(), excluded.read_bytes())
            catalog.write_catalog(books, records, [], output, excluded)
            self.assertEqual(first, (output.read_bytes(), excluded.read_bytes()))

    def test_duplicate_ids_rejected_before_write(self):
        with tempfile.TemporaryDirectory() as directory:
            output = pathlib.Path(directory) / 'catalog.json'
            with self.assertRaises(ValueError):
                catalog.write_catalog([], [{'id': 1}, {'id': 1}], [],
                                      output, pathlib.Path(directory) / 'quarantine.json')
            self.assertFalse(output.exists())


if __name__ == '__main__':
    unittest.main()

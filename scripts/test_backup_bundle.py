import io
import json
from pathlib import Path
import tarfile
import tempfile
import unittest

from backup_bundle import create, verify


class RecoveryBundleTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.root = Path(self.temp.name)
        self.stage = self.root / "stage"
        (self.stage / "attachments" / "nested").mkdir(parents=True)
        (self.stage / "database.dump").write_bytes(b"test pg archive")
        (self.stage / "attachments" / "nested" / "invoice.txt").write_text("invoice content")
        (self.stage / "attachments" / "manifest.json").write_text("document, not metadata")
        self.archive = self.root / "bundle.tar.gz"
        self.output = self.root / "output"
        self.output.mkdir()
        create(self.stage, self.archive, "erp_test")

    def tearDown(self):
        self.temp.cleanup()

    def test_roundtrip_preserves_documents_named_manifest(self):
        verify(self.archive, self.output)
        self.assertEqual((self.output / "attachments" / "manifest.json").read_text(), "document, not metadata")

    def rewrite(self, transform):
        replacement = self.root / "bad.tar.gz"
        with tarfile.open(self.archive) as source, tarfile.open(replacement, "w:gz") as target:
            for member in source.getmembers():
                content = source.extractfile(member).read() if member.isfile() else None
                member, content = transform(member, content)
                if content is not None:
                    member.size = len(content)
                target.addfile(member, io.BytesIO(content) if content is not None else None)
        return replacement

    def test_modified_attachment_is_rejected(self):
        bad = self.rewrite(lambda member, data: (member, b"tampered") if member.name.endswith("invoice.txt") else (member, data))
        with self.assertRaisesRegex(ValueError, "checksum mismatch"):
            verify(bad, self.output)

    def test_path_traversal_and_links_are_rejected_before_extraction(self):
        for name, type in [("../escape", tarfile.REGTYPE), ("attachments/link", tarfile.SYMTYPE)]:
            with self.subTest(name=name):
                bad = self.root / "malicious.tar.gz"
                with tarfile.open(bad, "w:gz") as archive:
                    member = tarfile.TarInfo(name)
                    member.type = type
                    member.linkname = "/etc/passwd"
                    archive.addfile(member, io.BytesIO(b""))
                with self.assertRaisesRegex(ValueError, "unsafe"):
                    verify(bad, self.output)
                self.assertEqual(list(self.output.iterdir()), [])

    def test_missing_document_is_rejected(self):
        def rename(member, data):
            if member.name.endswith("invoice.txt"):
                member.name = "attachments/other.txt"
            return member, data
        with self.assertRaisesRegex(ValueError, "inventory"):
            verify(self.rewrite(rename), self.output)


if __name__ == "__main__":
    unittest.main()

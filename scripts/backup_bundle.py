#!/usr/bin/env python3
"""Create/verify ERP database and attachment recovery bundles (format version 1)."""
import argparse
import hashlib
import json
from pathlib import Path, PurePosixPath
import tarfile
from datetime import datetime, timezone


def digest(path):
    with path.open("rb") as handle:
        return hashlib.file_digest(handle, "sha256").hexdigest()


def create(stage, archive, database):
    files = {}
    for path in sorted(stage.rglob("*")):
        if path.is_symlink():
            raise ValueError("symbolic links are not supported in document storage")
        if path.is_file():
            files[path.relative_to(stage).as_posix()] = {"sha256": digest(path), "bytes": path.stat().st_size}
    if "database.dump" not in files or not (stage / "attachments").is_dir():
        raise ValueError("database and attachments are both required")
    manifest = {"format": "erp-recovery-bundle", "version": 1, "database": database,
                "created_at": datetime.now(timezone.utc).isoformat(),
                "consistency": "writers-quiesced", "files": files}
    (stage / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n", encoding="utf-8")
    with tarfile.open(archive, "w:gz") as target:
        for name in ["manifest.json", "database.dump", "attachments"]:
            target.add(stage / name, arcname=name)


def verify(archive, stage):
    if any(stage.iterdir()):
        raise ValueError("verification destination must be empty")
    names = set()
    with tarfile.open(archive, "r:gz") as source:
        for member in source.getmembers():
            path = PurePosixPath(member.name)
            if (path.is_absolute() or ".." in path.parts or "\\" in member.name
                    or not path.parts or path.parts[0] not in {"manifest.json", "database.dump", "attachments"}
                    or not (member.isfile() or member.isdir()) or member.name in names):
                raise ValueError("unsafe or duplicate archive member: " + member.name)
            names.add(member.name)
        source.extractall(stage, filter="data")
    manifest = json.loads((stage / "manifest.json").read_text(encoding="utf-8"))
    if manifest.get("format") != "erp-recovery-bundle" or manifest.get("version") != 1 or manifest.get("consistency") != "writers-quiesced":
        raise ValueError("unsupported recovery manifest")
    expected = manifest.get("files", {})
    actual = {path.relative_to(stage).as_posix() for path in stage.rglob("*") if path.is_file() and path.relative_to(stage).as_posix() != "manifest.json"}
    if actual != set(expected) or "database.dump" not in actual or not (stage / "attachments").is_dir():
        raise ValueError("manifest does not match archive inventory")
    for name, metadata in expected.items():
        path = stage / name
        if path.stat().st_size != metadata["bytes"] or digest(path) != metadata["sha256"]:
            raise ValueError("checksum mismatch: " + name)
    print(json.dumps({"verified": True, "database": manifest["database"], "files": len(expected), "created_at": manifest["created_at"]}))


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("action", choices=["create", "verify"])
    parser.add_argument("stage", type=Path)
    parser.add_argument("archive", type=Path)
    parser.add_argument("--database", default="")
    args = parser.parse_args()
    if args.action == "create":
        create(args.stage, args.archive, args.database)
    else:
        verify(args.archive, args.stage)


if __name__ == "__main__":
    main()

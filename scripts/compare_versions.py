#!/usr/bin/env python3
import json
import sys

def load_files(filepath):
    with open(filepath, 'r') as f:
        data = json.load(f)
    # The gcloud output has a 'files' key where keys are filenames
    # and values are objects containing 'sha1Sum' and potentially other info.
    # Note: 'gcloud' describe doesn't always provide size directly, 
    # but it provides the list of files.
    return data.get('deployment', {}).get('files', {})

def main():
    if len(sys.argv) != 3:
        print("Usage: compare_versions.py version1.json version2.json")
        print("Generate json via: gcloud app versions describe <id> --format=json > version.json")
        sys.exit(1)

    v1_files = load_files(sys.argv[1])
    v2_files = load_files(sys.argv[2])

    v1_set = set(v1_files.keys())
    v2_set = set(v2_files.keys())

    added = v2_set - v1_set
    removed = v1_set - v2_set
    common = v1_set & v2_set

    print(f"Comparison of {sys.argv[1]} vs {sys.argv[2]}")
    print("-" * 40)
    print(f"Files in v1: {len(v1_set)}")
    print(f"Files in v2: {len(v2_set)}")
    print("-" * 40)

    if added:
        print(f"\nAdded in v2 ({len(added)} files):")
        for f in sorted(added):
            print(f"  + {f}")

    if removed:
        print(f"\nRemoved in v2 ({len(removed)} files):")
        for f in sorted(removed):
            print(f"  - {f}")

    # Note: Description doesn't include file size, only SHA.
    changed = [f for f in common if v1_files[f].get('sha1Sum') != v2_files[f].get('sha1Sum')]
    if changed:
        print(f"\nModified ({len(changed)} files):")
        for f in sorted(changed):
            print(f"  * {f}")

if __name__ == "__main__":
    main()

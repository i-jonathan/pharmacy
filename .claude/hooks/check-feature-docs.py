#!/usr/bin/env python3
"""PreToolUse hook: before `gh pr create`, require feature docs to be refreshed.

For each feature in docs/features/feature-paths.json whose code changed on this branch, the
branch must also change docs/features/<slug>/, or carry a commit trailer
`Feature-Docs-Unchanged: <slug>[, <slug>...]` recording that the owner checked the docs are
still accurate. Otherwise the PR is denied with instructions. Any failure of the hook itself
lets the command through, so a broken hook never blocks work.
"""
import fnmatch
import json
import os
import re
import shlex
import subprocess
import sys

MAPPING = "docs/features/feature-paths.json"
TRAILER = re.compile(r"^Feature-Docs-Unchanged:\s*(.+)$", re.MULTILINE)


def git(root, *args):
    return subprocess.run(
        ["git", "-C", root, *args], capture_output=True, text=True, check=True
    ).stdout


def base_branch(command):
    try:
        words = shlex.split(command)
    except ValueError:
        words = command.split()
    for i, word in enumerate(words):
        if word in ("--base", "-B") and i + 1 < len(words):
            return words[i + 1]
        if word.startswith("--base="):
            return word.split("=", 1)[1]
    return "main"


def main():
    event = json.load(sys.stdin)
    command = (event.get("tool_input") or {}).get("command", "")
    if not re.search(r"\bgh\s+pr\s+create\b", command):
        return

    root = os.environ.get("CLAUDE_PROJECT_DIR") or event.get("cwd") or "."
    mapping_path = os.path.join(root, MAPPING)
    if not os.path.exists(mapping_path):
        return
    with open(mapping_path) as f:
        mapping = json.load(f)

    base = base_branch(command)
    ref = f"origin/{base}"
    try:
        git(root, "rev-parse", "--verify", "--quiet", ref)
    except subprocess.CalledProcessError:
        ref = base
    # FEATURE_DOCS_BASE / FEATURE_DOCS_HEAD check an existing commit range, for testing the hook.
    head = os.environ.get("FEATURE_DOCS_HEAD", "HEAD")
    merge_base = os.environ.get("FEATURE_DOCS_BASE") or git(root, "merge-base", ref, head).strip()
    changed = git(root, "diff", "--name-only", f"{merge_base}..{head}").split()
    messages = git(root, "log", "--format=%B", f"{merge_base}..{head}")

    unchanged_ok = set()
    for line in TRAILER.findall(messages):
        unchanged_ok.update(s.strip() for s in line.split(",") if s.strip())

    stale = []
    for slug, patterns in mapping.items():
        code_hits = [p for p in changed if any(fnmatch.fnmatch(p, pat) for pat in patterns)]
        if not code_hits:
            continue
        docs_touched = any(p.startswith(f"docs/features/{slug}/") for p in changed)
        if docs_touched or slug in unchanged_ok:
            continue
        stale.append((slug, code_hits))

    if not stale:
        return

    lines = ["This branch changes code for features whose docs were not updated:"]
    for slug, hits in stale:
        shown = ", ".join(hits[:5]) + (f" and {len(hits) - 5} more" if len(hits) > 5 else "")
        lines.append(f"- {slug}: {shown}")
    lines += [
        "",
        "Before opening the PR, either:",
        (
            "1. refresh docs/features/<slug>/ with the document-feature skill (update mode) and "
            "commit it with the owner's approval, or"
        ),
        (
            "2. if the owner confirms the docs are still accurate, add a commit with the trailer "
            "`Feature-Docs-Unchanged: <slug>` (owner approval required for the commit)."
        ),
        "Then run gh pr create again.",
    ]
    json.dump(
        {
            "hookSpecificOutput": {
                "hookEventName": "PreToolUse",
                "permissionDecision": "deny",
                "permissionDecisionReason": "\n".join(lines),
            }
        },
        sys.stdout,
    )


if __name__ == "__main__":
    try:
        main()
    # Never block work because the hook itself failed: bad hook input or mapping (ValueError covers
    # JSON errors), git failing (CalledProcessError), or files that can't be read (OSError).
    except (ValueError, KeyError, TypeError, AttributeError, OSError,
            subprocess.CalledProcessError) as e:
        print(f"check-feature-docs hook skipped: {e}", file=sys.stderr)
    sys.exit(0)

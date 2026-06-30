#!/usr/bin/env python3
"""Сводный отчёт по результатам schemathesis-прогона.

Парсит junit-xml для трёх сервисов и печатает таблицу:
    эндпоинт, статус, число failure'ов.

Эта таблица — основа таблицы 5.1 из отчёта по фаззингу.
"""
from __future__ import annotations
import sys
import os
import re
import xml.etree.ElementTree as ET
from pathlib import Path


def parse_junit(path: Path) -> list[tuple[str, str, int, int]]:
    """Возвращает [(endpoint, status, total, failures), ...]"""
    if not path.is_file():
        return []
    tree = ET.parse(path)
    root = tree.getroot()
    rows = []
    for tc in root.iter("testcase"):
        name = tc.attrib.get("name", "")
        # У schemathesis имена вида "GET /groups", "POST /auth/register".
        # Если в одном testcase несколько проверок, считаем общее count.
        failures = len(list(tc.iter("failure"))) + len(list(tc.iter("error")))
        # max-examples в нашем run = переменная окружения N (см. run.sh),
        # но точное число попыток лучше взять из system-out, если есть.
        examples = _extract_examples_count(tc)
        status = "FAIL" if failures > 0 else "PASS"
        rows.append((name, status, examples, failures))
    return rows


def _extract_examples_count(tc) -> int:
    """Достаём 'X test cases' из system-out."""
    for child in tc:
        if child.tag.endswith("system-out") and child.text:
            m = re.search(r"(\d+)\s+test cases", child.text)
            if m:
                return int(m.group(1))
    # fallback: переменная окружения N
    return int(os.environ.get("N", "100"))


def main():
    out_dir = Path(sys.argv[1])
    services = ["auth", "groups", "projects"]
    print(f"# Сводка фаззинга — {out_dir.name}")
    print(f"{'#':>3} {'эндпоинт':<55} {'запросов':>9} {'аномалий':>9} {'статус':>10}")
    print("-" * 95)
    total_req, total_fail, n = 0, 0, 1
    for svc in services:
        rows = parse_junit(out_dir / f"{svc}.junit.xml")
        if not rows:
            continue
        for endpoint, status, examples, failures in rows:
            print(f"{n:>3} {svc + ' / ' + endpoint:<55} {examples:>9} {failures:>9} {status:>10}")
            total_req += examples
            total_fail += failures
            n += 1
    print("-" * 95)
    print(f"  Всего: {total_req} запросов, {total_fail} аномалий")


if __name__ == "__main__":
    main()

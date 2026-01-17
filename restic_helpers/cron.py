"""Cron expression handling and conversion"""

import math
from itertools import product

from croniter import croniter


class CronConversionError(Exception):
    pass


class InvalidCronExpression(CronConversionError):
    pass


class ScheduleEmpty(CronConversionError):
    pass


class ScheduleTooComplex(CronConversionError):
    pass


def cron_to_launchd(cron: croniter, max_entries=50):
    """Convert croniter object to launchd StartCalendarInterval"""
    expanded = cron.expanded

    keys = ["Minute", "Hour", "Day", "Month", "Weekday"]

    fields = {}
    for i, key in enumerate(keys):
        if expanded[i] == ["*"]:
            continue
        fields[key] = list(expanded[i])

    if not fields:
        raise ScheduleEmpty("Schedule has no constraints (would run every minute)")

    count = math.prod(len(v) for v in fields.values())
    if count > max_entries:
        raise ScheduleTooComplex(
            f"Schedule produces {count} entries (max {max_entries})"
        )

    keys = list(fields.keys())
    values = [fields[k] for k in keys]
    entries = [dict(zip(keys, combo)) for combo in product(*values)]

    return entries[0] if len(entries) == 1 else entries

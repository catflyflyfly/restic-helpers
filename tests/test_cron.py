import pytest
from croniter import croniter
from restic_helpers.cron import cron_to_launchd, ScheduleEmpty, ScheduleTooComplex

SINGLE_ENTRY_CASES = [
    pytest.param("30 14 * * *", {"Minute": 30, "Hour": 14}, id="daily_at_1430"),
    pytest.param("0 * * * *", {"Minute": 0}, id="hourly"),
    pytest.param(
        "0 9 * * 1", {"Minute": 0, "Hour": 9, "Weekday": 1}, id="weekly_monday"
    ),
    pytest.param("0 0 1 * *", {"Minute": 0, "Hour": 0, "Day": 1}, id="monthly_first"),
]

MULTI_ENTRY_CASES = [
    pytest.param(
        "0 9,17 * * *",
        [
            {"Minute": 0, "Hour": 9},
            {"Minute": 0, "Hour": 17},
        ],
        id="twice_daily",
    ),
    pytest.param(
        "30 14 * * 1,3,5",
        [
            {"Minute": 30, "Hour": 14, "Weekday": 1},
            {"Minute": 30, "Hour": 14, "Weekday": 3},
            {"Minute": 30, "Hour": 14, "Weekday": 5},
        ],
        id="three_weekdays",
    ),
]

ERROR_CASES = [
    pytest.param("* * * * *", ScheduleEmpty, id="all_wildcards"),
    pytest.param("*/5 */2 * * 1-5", ScheduleTooComplex, id="too_complex"),
]


@pytest.mark.parametrize("cron_expr,expected", SINGLE_ENTRY_CASES)
def test_single_entry(cron_expr, expected):
    assert cron_to_launchd(croniter(cron_expr)) == expected


@pytest.mark.parametrize("cron_expr,expected", MULTI_ENTRY_CASES)
def test_multi_entry(cron_expr, expected):
    assert cron_to_launchd(croniter(cron_expr)) == expected


@pytest.mark.parametrize("cron_expr,error", ERROR_CASES)
def test_errors(cron_expr, error):
    with pytest.raises(error):
        cron_to_launchd(croniter(cron_expr))


def test_max_entries_param():
    c = croniter("0 9,10,11 * * *")
    assert len(cron_to_launchd(c, max_entries=3)) == 3
    with pytest.raises(ScheduleTooComplex):
        cron_to_launchd(c, max_entries=2)

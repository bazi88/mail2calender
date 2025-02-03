import re
from datetime import datetime, timedelta
from typing import Optional, Dict, Any
from dateutil.rrule import rrule, DAILY, WEEKLY, MONTHLY, YEARLY
import pytz


class RecurringEventProcessor:
    """Process recurring event patterns in text"""

    def __init__(self, timezone: str = "Asia/Ho_Chi_Minh"):
        self.timezone = pytz.timezone(timezone)

        # Patterns for recurring events in Vietnamese
        self.recurring_patterns = {
            "daily": {"pattern": r"(mỗi|hàng)\s+ngày", "freq": DAILY},
            "weekly": {"pattern": r"(mỗi|hàng)\s+tuần", "freq": WEEKLY},
            "monthly": {"pattern": r"(mỗi|hàng)\s+tháng", "freq": MONTHLY},
            "yearly": {"pattern": r"(mỗi|hàng)\s+năm", "freq": YEARLY},
        }

        # Weekday patterns
        self.weekday_patterns = {
            "monday": r"thứ\s+hai|monday",
            "tuesday": r"thứ\s+ba|tuesday",
            "wednesday": r"thứ\s+tư|wednesday",
            "thursday": r"thứ\s+năm|thursday",
            "friday": r"thứ\s+sáu|friday",
            "saturday": r"thứ\s+bảy|saturday",
            "sunday": r"chủ\s+nhật|sunday",
        }

    def detect_recurrence(self, text: str) -> Optional[Dict[str, Any]]:
        """
        Detect recurring event pattern in text

        Args:
            text: Input text to analyze

        Returns:
            Optional[Dict]: Recurrence information if found
        """
        text = text.lower()

        for freq_type, freq_info in self.recurring_patterns.items():
            if re.search(freq_info["pattern"], text):
                recurrence = {
                    "type": freq_type,
                    "frequency": freq_info["freq"],
                }

                # Check for weekday specifications
                for weekday, pattern in self.weekday_patterns.items():
                    if re.search(pattern, text):
                        recurrence["weekday"] = weekday

                return recurrence
        return None

    def generate_rrule(
        self, recurrence: Dict[str, Any], start_date: datetime
    ) -> Optional[str]:
        """
        Generate RFC 5545 RRULE string

        Args:
            recurrence: Recurrence pattern information
            start_date: Start date for recurrence

        Returns:
            Optional[str]: RRULE string if valid
        """
        try:
            if not start_date.tzinfo:
                start_date = self.timezone.localize(start_date)

            rule_kwargs = {"freq": recurrence["frequency"], "dtstart": start_date}

            # Add weekday if specified
            if "weekday" in recurrence:
                rule_kwargs["byweekday"] = getattr(rrule, recurrence["weekday"].upper())

            # Create rule
            rule = rrule(**rule_kwargs)

            return rule._str().replace("RRULE:", "")

        except Exception as e:
            logger.error(f"Error generating RRULE: {e}")
            return None

    def get_next_occurrences(self, rrule_str: str, count: int = 5) -> list[datetime]:
        """
        Get next N occurrences of a recurring event

        Args:
            rrule_str: RRULE string
            count: Number of occurrences to return

        Returns:
            list[datetime]: Next occurrence dates
        """
        try:
            rule = rrule.rrulestr(f"RRULE:{rrule_str}")
            return list(rule[:count])

        except Exception as e:
            logger.error(f"Error getting next occurrences: {e}")
            return []

    def process_recurring_time(self, entity: Dict[str, Any]) -> Dict[str, Any]:
        """
        Process time entity for recurrence

        Args:
            entity: Time entity to process

        Returns:
            Dict: Processed entity with recurrence info
        """
        if "normalized_time" not in entity:
            return entity

        recurrence = self.detect_recurrence(entity["text"])
        if recurrence:
            start_date = datetime.fromisoformat(entity["normalized_time"])
            rrule_str = self.generate_rrule(recurrence, start_date)

            if rrule_str:
                entity["recurrence"] = {
                    "type": recurrence["type"],
                    "rrule": rrule_str,
                    "next_occurrences": [
                        dt.isoformat() for dt in self.get_next_occurrences(rrule_str)
                    ],
                }

        return entity

import re
from datetime import datetime, timedelta
from dateutil import parser
import pytz
from typing import Dict, Optional, Union


class TimeProcessor:
    """Process and normalize time expressions extracted by NER."""

    def __init__(self, timezone: str = "Asia/Ho_Chi_Minh"):
        self.timezone = pytz.timezone(timezone)
        # Common time patterns in Vietnamese
        self.time_patterns = {
            "today": r"hôm nay|ngày này",
            "tomorrow": r"ngày mai|hôm sau",
            "next_week": r"tuần sau|tuần tới",
            "next_month": r"tháng sau|tháng tới",
            "day_time": r"(\d{1,2})[h:giờ](\d{0,2})",
            "date_time": r"(\d{1,2})/(\d{1,2})(?:/(\d{4}|\d{2}))?",
        }

    def process_time_entity(self, entity: Dict) -> Optional[Dict]:
        """
        Process a time entity and normalize it to ISO format.

        Args:
            entity (Dict): Entity with text and type from NER model

        Returns:
            Dict: Processed entity with normalized time information
        """
        if entity["type"] not in ["TIME", "DATE"]:
            return entity

        text = entity["text"].lower()
        now = datetime.now(self.timezone)

        try:
            # Process relative time expressions
            if re.search(self.time_patterns["today"], text):
                normalized_time = now
            elif re.search(self.time_patterns["tomorrow"], text):
                normalized_time = now + timedelta(days=1)
            elif re.search(self.time_patterns["next_week"], text):
                normalized_time = now + timedelta(weeks=1)
            elif re.search(self.time_patterns["next_month"], text):
                # Add one month approximately
                normalized_time = now + timedelta(days=30)

            # Process time expressions (e.g., "15h30")
            elif re.search(self.time_patterns["day_time"], text):
                match = re.search(self.time_patterns["day_time"], text)
                hour = int(match.group(1))
                minute = int(match.group(2)) if match.group(2) else 0
                normalized_time = now.replace(hour=hour, minute=minute)

            # Process date expressions (e.g., "25/12/2023")
            elif re.search(self.time_patterns["date_time"], text):
                match = re.search(self.time_patterns["date_time"], text)
                day = int(match.group(1))
                month = int(match.group(2))
                year = int(match.group(3)) if match.group(3) else now.year
                # Handle 2-digit year
                if year < 100:
                    year += 2000
                normalized_time = now.replace(year=year, month=month, day=day)

            # Try general parsing for other formats
            else:
                normalized_time = parser.parse(text, fuzzy=True)
                if not normalized_time.tzinfo:
                    normalized_time = self.timezone.localize(normalized_time)

            # Update entity with normalized time
            entity.update(
                {
                    "normalized_time": normalized_time.isoformat(),
                    "timestamp": int(normalized_time.timestamp()),
                }
            )

            return entity

        except (ValueError, AttributeError) as e:
            # If parsing fails, return original entity
            return entity

    def get_duration(
        self, start_time: Union[str, datetime], end_time: Union[str, datetime]
    ) -> Optional[int]:
        """
        Calculate duration between two times in minutes.

        Args:
            start_time: Start time (ISO format string or datetime object)
            end_time: End time (ISO format string or datetime object)

        Returns:
            int: Duration in minutes or None if calculation fails
        """
        try:
            # Convert strings to datetime if needed
            if isinstance(start_time, str):
                start_time = parser.parse(start_time)
            if isinstance(end_time, str):
                end_time = parser.parse(end_time)

            # Ensure timezone awareness
            if not start_time.tzinfo:
                start_time = self.timezone.localize(start_time)
            if not end_time.tzinfo:
                end_time = self.timezone.localize(end_time)

            duration = end_time - start_time
            return int(duration.total_seconds() / 60)

        except (ValueError, TypeError):
            return None

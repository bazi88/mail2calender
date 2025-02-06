import re
from datetime import datetime, timedelta
from typing import Optional, Tuple, List

class VietnameseTimeProcessor:
    def __init__(self):
        self.vn_weekdays = {
            'thứ hai': 0, 'thứ 2': 0, 't2': 0,
            'thứ ba': 1, 'thứ 3': 1, 't3': 1,
            'thứ tư': 2, 'thứ 4': 2, 't4': 2,
            'thứ năm': 3, 'thứ 5': 3, 't5': 3,
            'thứ sáu': 4, 'thứ 6': 4, 't6': 4,
            'thứ bảy': 5, 'thứ 7': 5, 't7': 5,
            'chủ nhật': 6, 'cn': 6
        }
        
        self.vn_months = {
            'tháng một': 1, 'tháng 1': 1, 'th1': 1,
            'tháng hai': 2, 'tháng 2': 2, 'th2': 2,
            'tháng ba': 3, 'tháng 3': 3, 'th3': 3,
            'tháng tư': 4, 'tháng 4': 4, 'th4': 4,
            'tháng năm': 5, 'tháng 5': 5, 'th5': 5,
            'tháng sáu': 6, 'tháng 6': 6, 'th6': 6,
            'tháng bảy': 7, 'tháng 7': 7, 'th7': 7,
            'tháng tám': 8, 'tháng 8': 8, 'th8': 8,
            'tháng chín': 9, 'tháng 9': 9, 'th9': 9,
            'tháng mười': 10, 'tháng 10': 10, 'th10': 10,
            'tháng mười một': 11, 'tháng 11': 11, 'th11': 11,
            'tháng mười hai': 12, 'tháng 12': 12, 'th12': 12
        }

        self.relative_days = {
            'hôm nay': 0,
            'ngày mai': 1,
            'ngày kia': 2,
            'ngày mốt': 2,
            'hôm qua': -1,
            'hôm kia': -2
        }

    def parse_relative_time(self, text: str) -> Optional[datetime]:
        """Parse Vietnamese relative time expressions"""
        text = text.lower()
        
        for expr, days in self.relative_days.items():
            if expr in text:
                return datetime.now() + timedelta(days=days)

        # Handle "tuần tới", "tuần sau"
        if 'tuần tới' in text or 'tuần sau' in text:
            return datetime.now() + timedelta(weeks=1)

        # Handle "tháng tới", "tháng sau"
        if 'tháng tới' in text or 'tháng sau' in text:
            current = datetime.now()
            if current.month == 12:
                return current.replace(year=current.year + 1, month=1)
            return current.replace(month=current.month + 1)

        return None

    def parse_weekday(self, text: str) -> Optional[datetime]:
        """Parse Vietnamese weekday expressions"""
        text = text.lower()
        
        for day_name, day_num in self.vn_weekdays.items():
            if day_name in text:
                today = datetime.now()
                current_weekday = today.weekday()
                days_ahead = day_num - current_weekday
                
                if days_ahead <= 0:  # If the day has passed this week
                    days_ahead += 7  # Move to next week
                    
                return today + timedelta(days=days_ahead)

        return None

    def parse_date(self, text: str) -> Optional[datetime]:
        """Parse Vietnamese date expressions"""
        text = text.lower()
        
        # Pattern for dd/mm/yyyy or dd-mm-yyyy
        date_pattern = r'(\d{1,2})[-/](\d{1,2})(?:[-/](\d{2,4}))?'
        match = re.search(date_pattern, text)
        
        if match:
            day, month, year = match.groups()
            day = int(day)
            month = int(month)
            
            if year:
                year = int(year)
                if year < 100:  # Handle two-digit years
                    year += 2000
            else:
                year = datetime.now().year

            try:
                return datetime(year, month, day)
            except ValueError:
                return None

        # Handle text month names
        for month_name, month_num in self.vn_months.items():
            if month_name in text:
                # Look for day number before or after month
                day_pattern = r'(?:ngày\s*)?([0-9]{1,2})'
                day_matches = re.finditer(day_pattern, text)
                
                for match in day_matches:
                    day = int(match.group(1))
                    try:
                        return datetime(datetime.now().year, month_num, day)
                    except ValueError:
                        continue

        return None

    def parse_time(self, text: str) -> Optional[Tuple[int, int]]:
        """Parse Vietnamese time expressions"""
        text = text.lower()
        
        # Pattern for HH:MM or HH giờ MM
        time_pattern = r'(\d{1,2})(?::|(\s*giờ\s*))(?:(\d{1,2})(?:\s*phút)?)?'
        match = re.search(time_pattern, text)
        
        if match:
            hour = int(match.group(1))
            minutes = match.group(3)
            minutes = int(minutes) if minutes else 0

            if 0 <= hour <= 23 and 0 <= minutes <= 59:
                return (hour, minutes)

        return None

    def extract_datetime(self, text: str) -> List[datetime]:
        """Extract all possible datetime references from Vietnamese text"""
        results = []
        
        # Try parsing relative time first
        relative_dt = self.parse_relative_time(text)
        if relative_dt:
            results.append(relative_dt)

        # Try parsing weekday
        weekday_dt = self.parse_weekday(text)
        if weekday_dt:
            results.append(weekday_dt)

        # Try parsing date
        date_dt = self.parse_date(text)
        if date_dt:
            # Look for time in the text
            time_tuple = self.parse_time(text)
            if time_tuple:
                hour, minute = time_tuple
                date_dt = date_dt.replace(hour=hour, minute=minute)
            results.append(date_dt)

        return results
# Mail2Calendar Project Context

## Project Purpose
Mail2Calendar is a service designed to automatically extract calendar events from email content using Natural Language Processing (NER) and create corresponding calendar entries.

## Key Problems Solved
- Automated extraction of event details from emails
- Intelligent parsing of dates, times, and locations
- Integration between email and calendar systems
- Reduction of manual calendar entry creation

## Main Functionality
1. Email Processing
   - SMTP integration for receiving emails
   - Email content parsing and extraction
   - NER service for entity recognition

2. Calendar Integration
   - Support for multiple calendar providers
   - Automated event creation
   - Conflict detection and resolution

3. User Management
   - Authentication and authorization
   - User preferences and settings
   - Email and calendar account management

## Integration Points

### SMTP Integration
- Email receiving and processing
- Support for various email formats
- Attachment handling

### Calendar APIs
- Google Calendar
- Other calendar service providers

### NER Service
- Custom NER model for event extraction
- Entity recognition for:
  - Dates and times
  - Locations
  - Event titles
  - Participants

### Authentication
- OAuth2 for calendar services
- JWT for API authentication
- Session management
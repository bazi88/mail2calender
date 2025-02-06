# Mail2Calendar Project Context

## Mục Đích và Mục Tiêu
- Tự động chuyển đổi email thành sự kiện lịch
- Tối ưu hóa quy trình quản lý lịch từ email
- Giảm thiểu thời gian xử lý thủ công

## Vấn Đề Chính
1. Xử lý Email
   - Kết nối SMTP/IMAP để đọc email
   - Phân tích nội dung email
   - Trích xuất thông tin sự kiện

2. Nhận Dạng Thông Tin (NER)
   - Trích xuất thời gian, địa điểm
   - Xác định người tham gia
   - Phân loại loại sự kiện

3. Tích Hợp Calendar
   - Hỗ trợ Google Calendar
   - Tạo và cập nhật sự kiện
   - Đồng bộ hai chiều

## Chức Năng Chính
1. Email Processing
   - Kết nối nhiều tài khoản email
   - Lọc và phân loại email
   - Xử lý attachments

2. Event Extraction
   - NER service cho tiếng Việt và Anh
   - Xác thực thông tin trích xuất
   - Xử lý các định dạng thời gian

3. Calendar Management  
   - Tạo sự kiện tự động
   - Gửi lời mời tham gia
   - Cập nhật và hủy sự kiện

## Điểm Tích Hợp
1. Email Services
   - SMTP/IMAP protocols
   - Gmail API
   - IMAP/SMTP Services
   - Email Processing Pipeline
   - Attachment Handling

2. Calendar Services
   - Google Calendar API
   - CalDAV protocol

3. NER Service
   - SpaCy models
   - Custom trained models
   - API endpoints
# Active Development Context

## Nhánh Phát Triển
- Main branch: main
- Development branch: develop
- Feature branches:
  - feature/ner-vietnamese
  - feature/calendar-sync
  - feature/email-attachments

## Phát Triển Đang Diễn Ra
1. NER Service
   - Tối ưu hóa mô hình tiếng Việt
   - Cải thiện batch processing
   - Monitoring và metrics

2. Calendar Integration
   - Calendar sync implementation
   - Đồng bộ hai chiều
   - Xử lý conflict và retry logic

3. Infrastructure
   - Redis cache optimization
   - MinIO storage configuration
   - Monitoring stack deployment

## Thay Đổi Gần Đây
1. NER Service
   - Redis rate limiting implementation
   - gRPC health checks
   - Batch processing setup

2. Infrastructure
   - Docker compose optimization
   - Redis memory limits
   - Health check configurations

3. Email Processing
   - MinIO attachment storage
   - Improved email parsing
   - Basic filtering implementation

## Các Bước Tiếp Theo
1. Immediate Tasks
   - Optimize NER batch processing
   - Complete Google Calendar integration
   - Deploy monitoring stack

2. This Week
   - Fine-tune Vietnamese model
   - Implement retry mechanisms
   - Enhance error handling

3. Next Sprint
   - Advanced email filtering
   - Calendar sync improvements
   - Performance optimization

## Review & Metrics
1. Code Quality
   - Linter warnings
   - Test coverage
   - Technical debt

2. Performance
   - API response times
   - Resource usage
   - Error rates

3. Development Velocity
   - Sprint completion
   - Bug resolution
   - Feature delivery
# NER Service

Service trích xuất thông tin có cấu trúc từ văn bản sử dụng Named Entity Recognition.

## Cấu Trúc

```
/model
  ner_model.py     # NER model implementation
  trainer.py       # Model training logic
  utils.py         # Helper functions

/api
  main.py          # FastAPI application
  routes.py        # API endpoints
  schemas.py       # Data models

/data
  training/        # Training data
  evaluation/      # Test data
  models/          # Saved models
```

## Tính Năng

### NER Model
- Hỗ trợ tiếng Việt và Anh
- Trích xuất:
  - Thời gian
  - Địa điểm
  - Người tham gia
  - Sự kiện

### API
- REST endpoints
- Batch processing
- Model management
- Health checks

### Training
- Custom training data
- Model evaluation
- Hyperparameter tuning
- Export models

## Sử Dụng

```bash
# Install dependencies
pip install -r requirements.txt

# Start API server
uvicorn api.main:app --reload

# Train model
python -m model.trainer

# Run tests
pytest
```

## API Endpoints

```
POST /api/v1/extract
- Extract entities from text

GET /api/v1/models
- List available models

POST /api/v1/train
- Train new model

GET /health
- Service health check
``` 
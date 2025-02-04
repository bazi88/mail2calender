from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import List, Optional
import spacy
from transformers import AutoTokenizer, AutoModelForTokenClassification
import torch
from concurrent.futures import ThreadPoolExecutor
import time

app = FastAPI()

# Load models
vi_model = AutoModelForTokenClassification.from_pretrained("vinai/phobert-base-ner")
vi_tokenizer = AutoTokenizer.from_pretrained("vinai/phobert-base")
en_model = spacy.load("en_core_web_trf")

# Thread pool for batch processing
executor = ThreadPoolExecutor(max_workers=4)

class TextRequest(BaseModel):
    text: str
    language: str = "vi"

class BatchRequest(BaseModel):
    texts: List[TextRequest]

class Entity(BaseModel):
    text: str
    label: str
    start: int
    end: int
    confidence: float

class ExtractResponse(BaseModel):
    entities: List[Entity]
    processing_time: float

@app.post("/api/v1/extract", response_model=ExtractResponse)
async def extract_entities(request: TextRequest):
    start_time = time.time()
    
    if request.language == "vi":
        entities = process_vietnamese(request.text)
    else:
        entities = process_english(request.text)
        
    return ExtractResponse(
        entities=entities,
        processing_time=time.time() - start_time
    )

@app.post("/api/v1/batch-extract")
async def batch_extract(request: BatchRequest):
    start_time = time.time()
    
    # Process texts in parallel
    futures = []
    for text_req in request.texts:
        if text_req.language == "vi":
            future = executor.submit(process_vietnamese, text_req.text)
        else:
            future = executor.submit(process_english, text_req.text)
        futures.append(future)
    
    # Collect results
    results = []
    for future in futures:
        try:
            result = future.result()
            results.append(result)
        except Exception as e:
            results.append({"error": str(e)})
    
    return {
        "results": results,
        "processing_time": time.time() - start_time
    }

def process_vietnamese(text: str) -> List[Entity]:
    # Tokenize and get predictions
    inputs = vi_tokenizer(text, return_tensors="pt", padding=True)
    outputs = vi_model(**inputs)
    predictions = torch.argmax(outputs.logits, dim=2)
    
    # Convert predictions to entities
    entities = []
    current_entity = None
    
    for i, (token, pred) in enumerate(zip(inputs.input_ids[0], predictions[0])):
        token_text = vi_tokenizer.decode([token])
        if pred > 0:  # Not O tag
            confidence = torch.softmax(outputs.logits[0][i], dim=0)[pred].item()
            if current_entity and current_entity["label"] == pred:
                current_entity["text"] += token_text
                current_entity["end"] = i
            else:
                if current_entity:
                    entities.append(Entity(**current_entity))
                current_entity = {
                    "text": token_text,
                    "label": pred.item(),
                    "start": i,
                    "end": i,
                    "confidence": confidence
                }
    
    if current_entity:
        entities.append(Entity(**current_entity))
    
    return entities

def process_english(text: str) -> List[Entity]:
    doc = en_model(text)
    return [
        Entity(
            text=ent.text,
            label=ent.label_,
            start=ent.start_char,
            end=ent.end_char,
            confidence=0.9  # SpaCy doesn't provide confidence scores
        )
        for ent in doc.ents
    ]

@app.get("/health")
def health_check():
    return {"status": "healthy"}

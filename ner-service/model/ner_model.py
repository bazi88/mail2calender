import torch
from transformers import AutoTokenizer, AutoModelForTokenClassification
from typing import List, Dict, Optional, Tuple
from functools import lru_cache
import logging
import time

# Setup logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class NERModel:
    def __init__(self):
        try:
            self.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
            logger.info(f"Using device: {self.device}")
            
            # Sử dụng model đa ngôn ngữ cho NER
            model_name = "Davlan/xlm-roberta-base-ner-hrl"
            self.tokenizer = AutoTokenizer.from_pretrained(model_name)
            self.model = AutoModelForTokenClassification.from_pretrained(model_name)
            self.model.to(self.device)
            
            self.id2label = self.model.config.id2label
            logger.info(f"Model labels: {self.id2label}")
            
            # Cache cho kết quả xử lý
            self.cache = {}
            
        except Exception as e:
            logger.error(f"Error initializing NER model: {e}")
            raise

    def extract_entities(self, text: str) -> Tuple[List[Dict], float]:
        try:
            start_time = time.time()
            
            # Kiểm tra cache
            if text in self.cache:
                logger.info("Using cached result")
                return self.cache[text]
            
            # Tokenize
            inputs = self.tokenizer(text, return_tensors="pt", padding=True, truncation=True)
            inputs = {k: v.to(self.device) for k, v in inputs.items()}
            
            # Inference
            with torch.no_grad():
                outputs = self.model(**inputs)
                
            # Lấy predictions
            predictions = torch.argmax(outputs.logits, dim=2)
            scores = torch.softmax(outputs.logits, dim=2)
            
            # Debug info
            logger.info(f"Raw predictions shape: {predictions.shape}")
            logger.info(f"Predictions: {predictions[0].tolist()}")
            
            # Chuyển tokens thành text
            tokens = self.tokenizer.convert_ids_to_tokens(inputs['input_ids'][0])
            logger.info(f"Tokens: {tokens}")
            
            # Xử lý kết quả
            entities = []
            current_entity: Optional[Dict] = None
            
            for idx, (token, pred_id) in enumerate(zip(tokens, predictions[0])):
                pred_label = self.id2label[pred_id.item()]
                confidence = scores[0][idx][pred_id].item()
                
                logger.info(f"Token: {token}, Label: {pred_label}, Confidence: {confidence:.2f}")
                
                if pred_label != "O":
                    # Bỏ qua special tokens
                    if token in [self.tokenizer.cls_token, self.tokenizer.sep_token, self.tokenizer.pad_token]:
                        continue
                        
                    # Xử lý B- tags (beginning of entity)
                    if pred_label.startswith("B-"):
                        if current_entity is not None:
                            entities.append(current_entity.copy())
                        current_entity = {
                            "text": token.replace('▁', ''),
                            "type": pred_label[2:],
                            "confidence": confidence,
                            "start_pos": idx,
                            "end_pos": idx
                        }
                    
                    # Xử lý I- tags (inside of entity)
                    elif pred_label.startswith("I-"):
                        if current_entity is not None and current_entity["type"] == pred_label[2:]:
                            current_entity["text"] += token.replace('▁', '')
                            current_entity["end_pos"] = idx
                            current_entity["confidence"] = (current_entity["confidence"] + confidence) / 2
                
                else:  # O tag
                    if current_entity is not None:
                        entities.append(current_entity.copy())
                        current_entity = None
            
            # Thêm entity cuối cùng nếu có
            if current_entity is not None:
                entities.append(current_entity.copy())
            
            process_time = time.time() - start_time
            logger.info(f"Processed text in {process_time:.2f}s. Found {len(entities)} entities.")
            
            # Cache kết quả
            self.cache[text] = (entities, process_time)
            return entities, process_time
            
        except Exception as e:
            logger.error(f"Error extracting entities: {e}")
            return [], 0.0

    def batch_extract_entities(self, texts: List[str], batch_size: int = 8) -> List[List[Dict]]:
        """
        Xử lý nhiều đoạn text cùng lúc.
        
        Args:
            texts (List[str]): Danh sách các đoạn text cần xử lý
            batch_size (int): Số lượng text xử lý trong một batch
            
        Returns:
            List[List[Dict]]: Danh sách kết quả cho mỗi text
        """
        try:
            results = []
            for i in range(0, len(texts), batch_size):
                batch_texts = texts[i:i + batch_size]
                
                # Tokenize batch
                inputs = self.tokenizer(
                    batch_texts,
                    padding=True,
                    truncation=True,
                    max_length=512,
                    return_tensors="pt"
                )
                inputs.to(self.device)

                # Model inference
                outputs = self.model(**inputs)
                probabilities = torch.softmax(outputs.logits, dim=2)
                predictions = torch.argmax(outputs.logits, dim=2)
                confidence_scores = torch.max(probabilities, dim=2).values

                # Process each text in batch
                batch_results = []
                for j, (preds, confs) in enumerate(zip(predictions, confidence_scores)):
                    tokens = self.tokenizer.convert_ids_to_tokens(inputs["input_ids"][j])
                    labels = [self.id2label[p.item()] for p in preds]
                    confidences = [c.item() for c in confs]
                    
                    entities = self._process_entities(tokens, labels, confidences, batch_texts[j])
                    batch_results.append(entities)
                
                results.extend(batch_results)
                
            return results
            
        except Exception as e:
            logger.error(f"Error processing batch: {e}")
            return [[] for _ in texts]

    def _process_entities(self, tokens: List[str], labels: List[str], 
                        confidences: List[float], original_text: str) -> List[Dict]:
        """
        Xử lý tokens và labels để tạo danh sách entities.
        
        Args:
            tokens (List[str]): Danh sách tokens
            labels (List[str]): Danh sách labels
            confidences (List[float]): Danh sách confidence scores
            original_text (str): Text gốc
            
        Returns:
            List[Dict]: Danh sách các entities được tìm thấy
        """
        entities = []
        current_entity: Optional[Dict] = None
        
        for i, (token, label, conf) in enumerate(zip(tokens, labels, confidences)):
            if label.startswith("B-"):
                if current_entity:
                    entities.append(current_entity)
                current_entity = {
                    "text": token,
                    "type": label[2:],
                    "confidence": conf,
                    "start_pos": i,
                    "end_pos": i
                }
            elif label.startswith("I-") and current_entity:
                if current_entity["type"] == label[2:]:  # Chỉ gộp nếu cùng type
                    current_entity["text"] += " " + token
                    current_entity["end_pos"] = i
                    current_entity["confidence"] = min(current_entity["confidence"], conf)
            elif label == "O" and current_entity:
                entities.append(current_entity)
                current_entity = None
        
        if current_entity:
            entities.append(current_entity)

        # Post-process entities
        entities = self._clean_entities(entities, original_text)
            
        return entities

    def _clean_entities(self, entities: List[Dict], original_text: str) -> List[Dict]:
        """
        Làm sạch và chuẩn hóa các entities.
        
        Args:
            entities (List[Dict]): Danh sách entities cần xử lý
            original_text (str): Text gốc
            
        Returns:
            List[Dict]: Danh sách entities đã được làm sạch
        """
        cleaned = []
        for entity in entities:
            # Bỏ qua entities có confidence quá thấp
            if entity["confidence"] < 0.5:
                continue
                
            # Làm sạch text
            text = entity["text"].strip()
            if text.startswith("##"):
                text = text[2:]
            
            # Cập nhật entity
            entity["text"] = text
            
            # Thêm vào kết quả nếu text không rỗng
            if text:
                cleaned.append(entity)
                
        return cleaned

    def clear_cache(self):
        """Xóa cache"""
        self.extract_entities.cache_clear()
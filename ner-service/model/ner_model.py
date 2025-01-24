import torch
from transformers import AutoTokenizer, AutoModelForTokenClassification
from typing import List, Dict
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
            
            self.tokenizer = AutoTokenizer.from_pretrained("vinai/phobert-base")
            self.model = AutoModelForTokenClassification.from_pretrained(
                "vinai/phobert-base", 
                num_labels=29
            )
            self.model.to(self.device)
            
            self.id2label = {
                0: "O", 1: "B-PER", 2: "I-PER", 3: "B-ORG", 4: "I-ORG",
                5: "B-LOC", 6: "I-LOC", 7: "B-MISC", 8: "I-MISC",
                9: "B-TIME", 10: "I-TIME", 11: "B-DATE", 12: "I-DATE",
                13: "B-MONEY", 14: "I-MONEY", 15: "B-PERCENT", 16: "I-PERCENT",
                17: "B-NUMBER", 18: "I-NUMBER", 19: "B-EMAIL", 20: "I-EMAIL",
                21: "B-URL", 22: "I-URL", 23: "B-PHONE", 24: "I-PHONE",
                25: "B-ADDRESS", 26: "I-ADDRESS", 27: "B-OTHER", 28: "I-OTHER"
            }
            
            # Cache cho kết quả xử lý
            self.cache = {}
            
        except Exception as e:
            logger.error(f"Error initializing NER model: {e}")
            raise

    @torch.no_grad()
    @lru_cache(maxsize=1000)
    def extract_entities(self, text: str) -> List[Dict]:
        """
        Trích xuất entities từ một đoạn text.
        
        Args:
            text (str): Đoạn text cần xử lý
            
        Returns:
            List[Dict]: Danh sách các entities được tìm thấy
        """
        try:
            start_time = time.time()
            
            # Tokenize input
            inputs = self.tokenizer(
                text,
                return_tensors="pt",
                padding=True,
                truncation=True,
                max_length=512  # Giới hạn độ dài input
            )
            inputs.to(self.device)

            # Model inference
            outputs = self.model(**inputs)
            
            # Tính softmax để lấy confidence scores
            probabilities = torch.softmax(outputs.logits, dim=2)
            predictions = torch.argmax(outputs.logits, dim=2)
            confidence_scores = torch.max(probabilities, dim=2).values

            # Convert tokens và labels
            tokens = self.tokenizer.convert_ids_to_tokens(inputs["input_ids"][0])
            labels = [self.id2label[p.item()] for p in predictions[0]]
            confidences = [score.item() for score in confidence_scores[0]]

            # Process entities
            entities = self._process_entities(tokens, labels, confidences, text)
            
            process_time = time.time() - start_time
            logger.info(f"Processed text in {process_time:.2f}s. Found {len(entities)} entities.")
            
            return entities
            
        except Exception as e:
            logger.error(f"Error processing text: {e}")
            return []

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
        current_entity = None
        
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
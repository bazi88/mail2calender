from typing import List, Dict, Optional, TypedDict
import torch
from transformers import AutoTokenizer, AutoModelForTokenClassification
from functools import lru_cache
import logging
import time
from utils.time_processor import TimeProcessor

# Setup logging
logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)


class Entity(TypedDict, total=False):
    """Type definition for entity dictionary"""

    text: str
    type: str
    confidence: float
    start_pos: int
    end_pos: int
    normalized_time: Optional[str]
    timestamp: Optional[int]


class NERModel:
    def __init__(self):
        try:
            self.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
            logger.info(f"Using device: {self.device}")

            self.tokenizer = AutoTokenizer.from_pretrained("vinai/phobert-base")
            self.model = AutoModelForTokenClassification.from_pretrained(
                "vinai/phobert-base", num_labels=29
            )
            self.model.to(self.device)

            self.id2label = {
                0: "O",
                1: "B-PER",
                2: "I-PER",
                3: "B-ORG",
                4: "I-ORG",
                5: "B-LOC",
                6: "I-LOC",
                7: "B-MISC",
                8: "I-MISC",
                9: "B-TIME",
                10: "I-TIME",
                11: "B-DATE",
                12: "I-DATE",
                13: "B-MONEY",
                14: "I-MONEY",
                15: "B-PERCENT",
                16: "I-PERCENT",
                17: "B-NUMBER",
                18: "I-NUMBER",
                19: "B-EMAIL",
                20: "I-EMAIL",
                21: "B-URL",
                22: "I-URL",
                23: "B-PHONE",
                24: "I-PHONE",
                25: "B-ADDRESS",
                26: "I-ADDRESS",
                27: "B-OTHER",
                28: "I-OTHER",
            }

            # Initialize time processor
            self.time_processor = TimeProcessor()

            # Cache for processed results
            self.cache = {}

        except Exception as e:
            logger.error(f"Error initializing NER model: {e}")
            raise

    @torch.no_grad()
    @lru_cache(maxsize=1000)
    def extract_entities(self, text: str) -> List[Entity]:
        """
        Extract entities from a text.

        Args:
            text (str): Text to process

        Returns:
            List[Entity]: List of found entities
        """
        try:
            start_time = time.time()

            # Tokenize input
            inputs = self.tokenizer(
                text, return_tensors="pt", padding=True, truncation=True, max_length=512
            )
            inputs.to(self.device)

            # Model inference
            outputs = self.model(**inputs)

            # Calculate softmax for confidence scores
            probabilities = torch.softmax(outputs.logits, dim=2)
            predictions = torch.argmax(outputs.logits, dim=2)
            confidence_scores = torch.max(probabilities, dim=2).values

            # Convert tokens and labels
            tokens = self.tokenizer.convert_ids_to_tokens(inputs["input_ids"][0])
            labels = [self.id2label[p.item()] for p in predictions[0]]
            confidences = [score.item() for score in confidence_scores[0]]

            # Process entities
            entities = self._process_entities(tokens, labels, confidences, text)

            process_time = time.time() - start_time
            logger.info(
                f"Processed text in {process_time:.2f}s. Found {len(entities)} entities."
            )

            return entities

        except Exception as e:
            logger.error(f"Error processing text: {e}")
            return []

    def batch_extract_entities(
        self, texts: List[str], batch_size: int = 8
    ) -> List[List[Entity]]:
        """
        Process multiple texts simultaneously.

        Args:
            texts (List[str]): List of texts to process
            batch_size (int): Number of texts to process in one batch

        Returns:
            List[List[Entity]]: List of results for each text
        """
        try:
            results = []
            for i in range(0, len(texts), batch_size):
                batch_texts = texts[i : i + batch_size]

                # Tokenize batch
                inputs = self.tokenizer(
                    batch_texts,
                    padding=True,
                    truncation=True,
                    max_length=512,
                    return_tensors="pt",
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
                    tokens = self.tokenizer.convert_ids_to_tokens(
                        inputs["input_ids"][j]
                    )
                    labels = [self.id2label[p.item()] for p in preds]
                    confidences = [c.item() for c in confs]

                    entities = self._process_entities(
                        tokens, labels, confidences, batch_texts[j]
                    )
                    batch_results.append(entities)

                results.extend(batch_results)

            return results

        except Exception as e:
            logger.error(f"Error processing batch: {e}")
            return [[] for _ in texts]

    def _process_entities(
        self,
        tokens: List[str],
        labels: List[str],
        confidences: List[float],
        original_text: str,
    ) -> List[Entity]:
        """
        Process tokens and labels to create entity list.

        Args:
            tokens (List[str]): List of tokens
            labels (List[str]): List of labels
            confidences (List[float]): List of confidence scores
            original_text (str): Original text

        Returns:
            List[Entity]: List of found entities
        """
        entities: List[Entity] = []
        current_entity: Optional[Entity] = None

        for i, (token, label, conf) in enumerate(zip(tokens, labels, confidences)):
            if label.startswith("B-"):
                if current_entity is not None:
                    entities.append(current_entity)
                current_entity = Entity(
                    text=token, type=label[2:], confidence=conf, start_pos=i, end_pos=i
                )
            elif label.startswith("I-") and current_entity is not None:
                if current_entity["type"] == label[2:]:  # Only merge if same type
                    current_entity = Entity(
                        text=f"{current_entity['text']} {token}",
                        type=current_entity["type"],
                        confidence=min(current_entity["confidence"], conf),
                        start_pos=current_entity["start_pos"],
                        end_pos=i,
                    )
            elif label == "O" and current_entity is not None:
                entities.append(current_entity)
                current_entity = None

        if current_entity is not None:
            entities.append(current_entity)

        # Clean and process time entities
        cleaned_entities: List[Entity] = []
        for entity in entities:
            # Skip low confidence entities
            if entity["confidence"] < 0.5:
                continue

            # Clean text
            text = entity["text"].strip()
            if text.startswith("##"):
                text = text[2:]

            # Create new entity with cleaned text
            clean_entity = Entity(
                text=text,
                type=entity["type"],
                confidence=entity["confidence"],
                start_pos=entity["start_pos"],
                end_pos=entity["end_pos"],
            )

            # Process time entities
            if clean_entity["type"] in ["TIME", "DATE"]:
                processed_entity = self.time_processor.process_time_entity(
                    dict(clean_entity)
                )
                if processed_entity and "normalized_time" in processed_entity:
                    cleaned_entities.append(Entity(**processed_entity))
            elif text:  # Add non-time entities if text is not empty
                cleaned_entities.append(clean_entity)

        return cleaned_entities

    def clear_cache(self):
        """Clear cache"""
        self.extract_entities.cache_clear()

"""
ML model for shipment priority prediction.

Trains a RandomForestClassifier on synthetic data and exposes
a predict() function that returns priority + per-factor contributions.
"""

import os
import pickle
import numpy as np
import pandas as pd
from sklearn.ensemble import RandomForestClassifier
from sklearn.model_selection import cross_val_score
from config import (
    PRIORITY_FACTORS,
    MODEL_FILE,
    RANDOM_STATE,
    N_ESTIMATORS,
)


# Feature columns used for training
FEATURE_COLS = [
    "norm_shipment_type",
    "norm_distance",
    "norm_restrictions",
    "norm_time_window",
    "norm_volume",
    "norm_saturation",
]

# Maps feature columns back to factor names
FEATURE_TO_FACTOR = {
    "norm_shipment_type": "shipment_type",
    "norm_distance": "distance_km",
    "norm_restrictions": "restrictions",
    "norm_time_window": "time_window",
    "norm_volume": "volume_score",
    "norm_saturation": "route_saturation",
}


def train_model(df: pd.DataFrame) -> RandomForestClassifier:
    """Train a RandomForestClassifier on the dataset."""
    X = df[FEATURE_COLS].values
    y = df["priority"].values

    clf = RandomForestClassifier(
        n_estimators=N_ESTIMATORS,
        random_state=RANDOM_STATE,
        max_depth=10,
        min_samples_split=10,
        min_samples_leaf=5,
    )
    clf.fit(X, y)

    # Cross-validation score
    scores = cross_val_score(clf, X, y, cv=5, scoring="accuracy")
    print(f"Cross-validation accuracy: {scores.mean():.3f} (+/- {scores.std():.3f})")

    return clf


def save_model(clf: RandomForestClassifier, path: str = MODEL_FILE):
    """Save the trained model to disk."""
    with open(path, "wb") as f:
        pickle.dump(clf, f)
    print(f"Model saved to {path}")


def load_model(path: str = MODEL_FILE) -> RandomForestClassifier:
    """Load a trained model from disk."""
    with open(path, "rb") as f:
        return pickle.load(f)


def predict(clf: RandomForestClassifier, normalized_factors: dict, raw_values: dict) -> dict:
    """
    Predict priority for a single shipment.

    Args:
        clf: trained classifier
        normalized_factors: dict of factor_name -> normalized float (0.0-1.0)
        raw_values: dict of factor_name -> original value (for display)

    Returns:
        dict with priority, confidence, score, and per-factor contributions
    """
    # Build feature vector in the correct order
    feature_names = list(FEATURE_TO_FACTOR.keys())
    x = np.array([[normalized_factors[FEATURE_TO_FACTOR[f]] for f in feature_names]])

    # Predict
    priority = clf.predict(x)[0]
    proba = clf.predict_proba(x)[0]
    classes = clf.classes_
    confidence = float(max(proba))

    # Compute score (weighted sum)
    total_weight = sum(cfg["weight"] for cfg in PRIORITY_FACTORS.values())
    weighted_sum = 0.0
    for factor_name, cfg in PRIORITY_FACTORS.items():
        weighted_sum += normalized_factors[factor_name] * cfg["weight"]
    score = weighted_sum / total_weight

    # Compute per-factor contributions
    factors = {}
    for factor_name, cfg in PRIORITY_FACTORS.items():
        normalized = normalized_factors[factor_name]
        weight = cfg["weight"]
        contribution = (normalized * weight) / total_weight
        factors[factor_name] = {
            "value": raw_values.get(factor_name, ""),
            "normalized": round(normalized, 4),
            "weight": weight,
            "contribution": round(contribution, 4),
        }

    return {
        "priority": priority,
        "confidence": round(confidence, 4),
        "score": round(score, 4),
        "factors": factors,
    }

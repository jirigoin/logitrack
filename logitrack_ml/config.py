"""
PRIORITY FACTORS — Ranked by importance (highest weight = most impact on priority).

To change which factors matter most:
  1. Adjust the weights below
  2. Run: python train.py
  3. Restart the ML service

The scoring formula is: total_score = Σ(normalized_factor × weight) / Σ(weights)
Then: alta if score > ALTA_THRESHOLD, media if > MEDIA_THRESHOLD, else baja.
"""

# Factor weights (higher = more influence on final priority)
PRIORITY_FACTORS = {
    "shipment_type":    {"weight": 3.0, "description": "Express = highest priority"},
    "distance_km":      {"weight": 2.5, "description": "Longer distance = more delay risk"},
    "restrictions":     {"weight": 2.0, "description": "Fragile/cold chain = needs special handling"},
    "time_window":      {"weight": 1.5, "description": "Morning = tighter deadline"},
    "volume_score":     {"weight": 1.0, "description": "Larger = more complex logistics"},
    "route_saturation": {"weight": 0.8, "description": "Busy route = congestion risk"},
}

# Thresholds for classification
ALTA_THRESHOLD = 0.65   # score > 0.65 → alta
MEDIA_THRESHOLD = 0.35  # score > 0.35 → media, else baja

# Volume score: package type base sizes
PACKAGE_BASE_SIZE = {
    "envelope": 1,
    "box":      5,
    "pallet":  15,
}

# Dataset generation
DATASET_SIZE = 2000
LABEL_NOISE_RATE = 0.10  # 10% of labels are randomly flipped

# Model
MODEL_FILE = "model.pkl"
RANDOM_STATE = 42
N_ESTIMATORS = 100

"""
Synthetic dataset generator for shipment priority prediction.

Generates random samples with realistic distributions and labels them
using the weighted scoring formula from config.py.
"""

import random
import numpy as np
import pandas as pd
from config import (
    PRIORITY_FACTORS,
    ALTA_THRESHOLD,
    MEDIA_THRESHOLD,
    PACKAGE_BASE_SIZE,
    DATASET_SIZE,
    LABEL_NOISE_RATE,
    RANDOM_STATE,
)

# Province coordinates (lat, lng) — centroids for all 23 provinces + CABA
PROVINCE_COORDS = {
    "Buenos Aires":         (-36.60, -60.50),
    "Ciudad de Buenos Aires": (-34.60, -58.38),
    "Catamarca":            (-28.47, -65.78),
    "Chaco":                (-26.39, -60.73),
    "Chubut":               (-43.30, -68.90),
    "Córdoba":              (-31.40, -64.18),
    "Corrientes":           (-28.66, -58.44),
    "Entre Ríos":           (-32.00, -59.20),
    "Formosa":              (-25.18, -59.73),
    "Jujuy":                (-23.32, -65.73),
    "La Pampa":             (-36.62, -65.45),
    "La Rioja":             (-29.41, -66.85),
    "Mendoza":              (-33.88, -68.83),
    "Misiones":             (-26.88, -54.58),
    "Neuquén":              (-38.95, -68.06),
    "Río Negro":            (-40.30, -67.30),
    "Salta":                (-24.78, -65.42),
    "San Juan":             (-31.53, -68.52),
    "San Luis":             (-33.30, -66.34),
    "Santa Cruz":           (-48.80, -69.65),
    "Santa Fe":             (-31.00, -61.00),
    "Santiago del Estero":  (-27.78, -63.25),
    "Tierra del Fuego":     (-53.80, -67.70),
    "Tucumán":              (-26.82, -65.22),
}

PROVINCES = list(PROVINCE_COORDS.keys())
SHIPMENT_TYPES = ["normal", "express"]
TIME_WINDOWS = ["morning", "afternoon", "flexible"]
PACKAGE_TYPES = ["envelope", "box", "pallet"]


def haversine_km(lat1: float, lng1: float, lat2: float, lng2: float) -> float:
    """Calculate distance between two points on Earth using Haversine formula."""
    R = 6371  # Earth radius in km
    lat1, lng1, lat2, lng2 = map(np.radians, [lat1, lng1, lat2, lng2])
    dlat = lat2 - lat1
    dlng = lng2 - lng1
    a = np.sin(dlat / 2) ** 2 + np.cos(lat1) * np.cos(lat2) * np.sin(dlng / 2) ** 2
    c = 2 * np.arcsin(np.sqrt(a))
    return R * c


def compute_distance(origin_province: str, dest_province: str) -> float:
    """Compute distance in km between two provinces."""
    o = PROVINCE_COORDS.get(origin_province, PROVINCE_COORDS["Ciudad de Buenos Aires"])
    d = PROVINCE_COORDS.get(dest_province, PROVINCE_COORDS["Ciudad de Buenos Aires"])
    return haversine_km(o[0], o[1], d[0], d[1])


def normalize_factor(factor_name: str, value) -> float:
    """Normalize a factor value to 0.0-1.0 range."""
    if factor_name == "shipment_type":
        return 1.0 if value == "express" else 0.0
    elif factor_name == "distance_km":
        return min(value / 2500.0, 1.0)
    elif factor_name == "restrictions":
        return value / 2.0  # value is count (0, 1, or 2)
    elif factor_name == "time_window":
        return {"flexible": 0.0, "afternoon": 0.5, "morning": 1.0}[value]
    elif factor_name == "volume_score":
        return min(value / 25.0, 1.0)
    elif factor_name == "route_saturation":
        return value  # already 0.0-1.0
    return 0.0


def compute_volume_score(package_type: str, weight_kg: float) -> float:
    """Compute volume score from package type and weight."""
    base = PACKAGE_BASE_SIZE.get(package_type, 5)
    return base + (weight_kg / 2.0)


def label_priority(score: float) -> str:
    """Convert a 0-1 score to a priority label."""
    if score > ALTA_THRESHOLD:
        return "alta"
    elif score > MEDIA_THRESHOLD:
        return "media"
    else:
        return "baja"


def compute_score(factors: dict) -> float:
    """Compute weighted score from normalized factors."""
    total_weight = sum(cfg["weight"] for cfg in PRIORITY_FACTORS.values())
    weighted_sum = 0.0
    for factor_name, cfg in PRIORITY_FACTORS.items():
        normalized = factors[factor_name]
        weighted_sum += normalized * cfg["weight"]
    return weighted_sum / total_weight


def generate_dataset(size: int = DATASET_SIZE, seed: int = RANDOM_STATE) -> pd.DataFrame:
    """Generate synthetic dataset of shipment samples with priority labels."""
    rng = random.Random(seed)
    np_rng = np.random.RandomState(seed)

    rows = []
    for _ in range(size):
        origin = rng.choice(PROVINCES)
        dest = rng.choice(PROVINCES)
        distance = compute_distance(origin, dest)
        shipment_type = rng.choice(SHIPMENT_TYPES)
        time_window = rng.choice(TIME_WINDOWS)
        package_type = rng.choice(PACKAGE_TYPES)
        weight_kg = round(rng.uniform(0.1, 50.0), 1)
        is_fragile = rng.random() < 0.25  # 25% are fragile
        cold_chain = rng.random() < 0.10  # 10% need cold chain
        route_saturation = round(rng.uniform(0.0, 1.0), 2)

        restriction_count = int(is_fragile) + int(cold_chain)
        volume = compute_volume_score(package_type, weight_kg)

        # Compute normalized factors
        normalized = {
            "shipment_type":    normalize_factor("shipment_type", shipment_type),
            "distance_km":      normalize_factor("distance_km", distance),
            "restrictions":     normalize_factor("restrictions", restriction_count),
            "time_window":      normalize_factor("time_window", time_window),
            "volume_score":     normalize_factor("volume_score", volume),
            "route_saturation": normalize_factor("route_saturation", route_saturation),
        }

        score = compute_score(normalized)
        label = label_priority(score)

        rows.append({
            "origin_province": origin,
            "destination_province": dest,
            "distance_km": round(distance, 1),
            "shipment_type": shipment_type,
            "time_window": time_window,
            "package_type": package_type,
            "weight_kg": weight_kg,
            "is_fragile": int(is_fragile),
            "cold_chain": int(cold_chain),
            "restriction_count": restriction_count,
            "volume_score": round(volume, 2),
            "route_saturation": route_saturation,
            "norm_shipment_type": normalized["shipment_type"],
            "norm_distance": normalized["distance_km"],
            "norm_restrictions": normalized["restrictions"],
            "norm_time_window": normalized["time_window"],
            "norm_volume": normalized["volume_score"],
            "norm_saturation": normalized["route_saturation"],
            "score": round(score, 4),
            "priority": label,
        })

    df = pd.DataFrame(rows)

    # Apply label noise: flip some labels randomly
    noise_mask = np_rng.random(len(df)) < LABEL_NOISE_RATE
    noise_indices = df.index[noise_mask]
    for idx in noise_indices:
        current = df.loc[idx, "priority"]
        options = [p for p in ["alta", "media", "baja"] if p != current]
        df.loc[idx, "priority"] = rng.choice(options)

    return df


if __name__ == "__main__":
    df = generate_dataset()
    print(f"Generated {len(df)} samples")
    print(f"\nPriority distribution:")
    print(df["priority"].value_counts())
    print(f"\nSample rows:")
    print(df.head(10).to_string())

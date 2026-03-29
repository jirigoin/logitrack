"""
Flask server for shipment priority prediction.

POST /predict — returns priority classification for a shipment.
GET  /health  — health check.
"""

import os
from flask import Flask, request, jsonify
from flask_cors import CORS
from model import load_model, predict
from dataset import normalize_factor, compute_volume_score, compute_distance, PROVINCE_COORDS
from config import MODEL_FILE

app = Flask(__name__)
CORS(app)

# Load model on startup
clf = None
if os.path.exists(MODEL_FILE):
    try:
        clf = load_model(MODEL_FILE)
        print(f"Model loaded from {MODEL_FILE}")
    except Exception as e:
        print(f"WARNING: Failed to load model: {e}")
else:
    print(f"WARNING: {MODEL_FILE} not found. Run 'python train.py' first.")


@app.route("/predict", methods=["POST"])
def predict_priority():
    """Predict shipment priority based on shipment attributes."""
    if clf is None:
        return jsonify({"error": "Model not loaded. Run 'python train.py' first."}), 503

    data = request.get_json()
    if not data:
        return jsonify({"error": "Request body is required"}), 400

    # Extract and validate fields
    origin_province = data.get("origin_province", "Ciudad de Buenos Aires")
    destination_province = data.get("destination_province", "Ciudad de Buenos Aires")
    shipment_type = data.get("shipment_type", "normal")
    time_window = data.get("time_window", "flexible")
    package_type = data.get("package_type", "box")
    weight_kg = float(data.get("weight_kg", 1.0))
    is_fragile = bool(data.get("is_fragile", False))
    cold_chain = bool(data.get("cold_chain", False))

    # Validate enums
    if shipment_type not in ("normal", "express"):
        return jsonify({"error": "shipment_type must be 'normal' or 'express'"}), 400
    if time_window not in ("morning", "afternoon", "flexible"):
        return jsonify({"error": "time_window must be 'morning', 'afternoon', or 'flexible'"}), 400
    if package_type not in ("envelope", "box", "pallet"):
        return jsonify({"error": "package_type must be 'envelope', 'box', or 'pallet'"}), 400

    # Compute derived values
    distance = compute_distance(origin_province, destination_province)
    restriction_count = int(is_fragile) + int(cold_chain)
    volume = compute_volume_score(package_type, weight_kg)

    # Simulated route saturation (deterministic based on route hash)
    route_hash = hash(f"{origin_province}-{destination_province}") % 100
    route_saturation = round(route_hash / 100.0, 2)

    # Normalize factors
    normalized = {
        "shipment_type":    normalize_factor("shipment_type", shipment_type),
        "distance_km":      normalize_factor("distance_km", distance),
        "restrictions":     normalize_factor("restrictions", restriction_count),
        "time_window":      normalize_factor("time_window", time_window),
        "volume_score":     normalize_factor("volume_score", volume),
        "route_saturation": normalize_factor("route_saturation", route_saturation),
    }

    raw_values = {
        "shipment_type":    shipment_type,
        "distance_km":      round(distance, 1),
        "restrictions":     restriction_count,
        "time_window":      time_window,
        "volume_score":     round(volume, 2),
        "route_saturation": route_saturation,
    }

    result = predict(clf, normalized, raw_values)
    return jsonify(result), 200


@app.route("/health", methods=["GET"])
def health():
    """Health check endpoint."""
    return jsonify({"status": "ok"}), 200


if __name__ == "__main__":
    port = int(os.environ.get("ML_PORT", 5001))
    print(f"ML Prediction Service running on :{port}")
    app.run(host="0.0.0.0", port=port, debug=False)

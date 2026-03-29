"""
One-time training script.

Generates synthetic dataset and trains the RandomForest model.
Run: python train.py
"""

from dataset import generate_dataset
from model import train_model, save_model
from config import DATASET_SIZE


def main():
    print("=" * 60)
    print("LogiTrack ML — Shipment Priority Model Training")
    print("=" * 60)

    # Step 1: Generate dataset
    print(f"\n[1/3] Generating synthetic dataset ({DATASET_SIZE} samples)...")
    df = generate_dataset()
    print(f"  Priority distribution:")
    for priority in ["alta", "media", "baja"]:
        count = len(df[df["priority"] == priority])
        pct = count / len(df) * 100
        print(f"    {priority:>5}: {count:>5} ({pct:.1f}%)")

    # Step 2: Train model
    print(f"\n[2/3] Training RandomForest classifier...")
    clf = train_model(df)

    # Step 3: Save model
    print(f"\n[3/3] Saving model...")
    save_model(clf)

    # Quick validation
    print(f"\n--- Sample predictions ---")
    from model import predict
    from dataset import normalize_factor

    test_cases = [
        {
            "name": "Express long-distance fragile",
            "normalized": {
                "shipment_type": 1.0,
                "distance_km": normalize_factor("distance_km", 2200),
                "restrictions": 1.0,
                "time_window": 1.0,
                "volume_score": normalize_factor("volume_score", 10),
                "route_saturation": 0.8,
            },
            "raw": {
                "shipment_type": "express",
                "distance_km": 2200,
                "restrictions": 1,
                "time_window": "morning",
                "volume_score": 10,
                "route_saturation": 0.8,
            },
        },
        {
            "name": "Normal short-distance envelope",
            "normalized": {
                "shipment_type": 0.0,
                "distance_km": normalize_factor("distance_km", 200),
                "restrictions": 0.0,
                "time_window": 0.0,
                "volume_score": normalize_factor("volume_score", 2),
                "route_saturation": 0.2,
            },
            "raw": {
                "shipment_type": "normal",
                "distance_km": 200,
                "restrictions": 0,
                "time_window": "flexible",
                "volume_score": 2,
                "route_saturation": 0.2,
            },
        },
    ]

    for tc in test_cases:
        result = predict(clf, tc["normalized"], tc["raw"])
        print(f"\n  {tc['name']}:")
        print(f"    Priority: {result['priority']} (confidence: {result['confidence']:.2f}, score: {result['score']:.2f})")
        for factor, info in sorted(result["factors"].items(), key=lambda x: -x[1]["contribution"]):
            print(f"      {factor:>20}: contribution={info['contribution']:.3f} (weight={info['weight']}, normalized={info['normalized']})")

    print(f"\n{'=' * 60}")
    print("Training complete. Run 'python app.py' to start the service.")
    print("=" * 60)


if __name__ == "__main__":
    main()

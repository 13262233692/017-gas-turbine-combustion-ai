import torch
import numpy as np
from flask import Flask, request, jsonify
from flask_cors import CORS
import os
import sys
import logging

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from models.cnn_pde_model import CNNPDEModel, InstabilityDetector, EfficiencyPredictor

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = Flask(__name__)
CORS(app)

device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
logger.info(f"Inference server using device: {device}")

temp_model = CNNPDEModel(in_channels=4, grid_size=16).to(device)
instability_model = InstabilityDetector(input_size=8, hidden_size=64, num_layers=2).to(device)
efficiency_model = EfficiencyPredictor(input_dim=10).to(device)

model_dir = os.path.join(os.path.dirname(os.path.abspath(__file__)), "models")
temp_model_path = os.path.join(model_dir, "temperature_cnn_pde.pt")
instability_model_path = os.path.join(model_dir, "instability_detector.pt")

if os.path.exists(temp_model_path):
    temp_model.load_state_dict(torch.load(temp_model_path, map_location=device))
    logger.info("Loaded temperature model weights")
else:
    logger.warning("No temperature model weights found, using random initialization")

if os.path.exists(instability_model_path):
    instability_model.load_state_dict(torch.load(instability_model_path, map_location=device))
    logger.info("Loaded instability detector weights")
else:
    logger.warning("No instability detector weights found, using random initialization")

temp_model.eval()
instability_model.eval()
efficiency_model.eval()


def sensor_data_to_input(sensor_data):
    grid_size = 16
    temp_grid = np.full((grid_size, grid_size), 1200.0, dtype=np.float32)
    pressure_grid = np.full((grid_size, grid_size), 1.5, dtype=np.float32)
    flow_grid = np.full((grid_size, grid_size), 2.5, dtype=np.float32)
    quality_grid = np.ones((grid_size, grid_size), dtype=np.float32)

    temp_sensors = [s for s in sensor_data if s.get("type") == "temperature"]
    for i, s in enumerate(temp_sensors):
        angle = 2 * np.pi * i / max(len(temp_sensors), 1)
        r = 0.6
        si = int((r * np.cos(angle) + 1) / 2 * (grid_size - 1))
        sj = int((r * np.sin(angle) + 1) / 2 * (grid_size - 1))
        si = np.clip(si, 0, grid_size - 1)
        sj = np.clip(sj, 0, grid_size - 1)
        from scipy.ndimage import gaussian_filter
        for di in range(-2, 3):
            for dj in range(-2, 3):
                ni, nj = si + di, sj + dj
                if 0 <= ni < grid_size and 0 <= nj < grid_size:
                    dist = np.sqrt(di ** 2 + dj ** 2)
                    w = np.exp(-dist ** 2 / 2.0)
                    temp_grid[ni, nj] = temp_grid[ni, nj] * (1 - w) + s["value"] * w

    from scipy.ndimage import gaussian_filter
    temp_grid = gaussian_filter(temp_grid, sigma=2.0)

    input_tensor = np.stack([temp_grid, pressure_grid, flow_grid, quality_grid], axis=0)
    return torch.tensor(input_tensor).unsqueeze(0).to(device)


@app.route("/api/predict/temperature", methods=["POST"])
def predict_temperature():
    try:
        sensor_data = request.json.get("sensors", [])
        if not sensor_data:
            return jsonify({"error": "No sensor data provided"}), 400

        input_tensor = sensor_data_to_input(sensor_data)

        with torch.no_grad():
            prediction = temp_model(input_tensor)

        field = prediction.squeeze().cpu().numpy()
        result = {
            "grid": field.tolist(),
            "rows": field.shape[0],
            "cols": field.shape[1],
            "max_temp": float(field.max()),
            "min_temp": float(field.min()),
            "avg_temp": float(field.mean()),
        }
        return jsonify(result)

    except Exception as e:
        logger.error(f"Temperature prediction error: {e}")
        return jsonify({"error": str(e)}), 500


@app.route("/api/predict/instability", methods=["POST"])
def predict_instability():
    try:
        sensor_data = request.json.get("sensors", [])
        if not sensor_data:
            return jsonify({"error": "No sensor data provided"}), 400

        input_tensor = sensor_data_to_input(sensor_data)

        with torch.no_grad():
            temp_pred = temp_model(input_tensor)
            features = extract_features(input_tensor, temp_pred)
            features_seq = features.unsqueeze(1).repeat(1, 10, 1)
            instability_prob = instability_model(features_seq).squeeze()

        prob = float(instability_prob.cpu())
        result = {
            "stable": prob < 0.35,
            "instability_index": prob,
            "risk_level": "critical" if prob > 0.7 else "warning" if prob > 0.35 else "normal",
        }
        return jsonify(result)

    except Exception as e:
        logger.error(f"Instability prediction error: {e}")
        return jsonify({"error": str(e)}), 500


@app.route("/api/predict/efficiency", methods=["POST"])
def predict_efficiency():
    try:
        sensor_data = request.json.get("sensors", [])
        if not sensor_data:
            return jsonify({"error": "No sensor data provided"}), 400

        temp_vals = [s["value"] for s in sensor_data if s.get("type") == "temperature"]
        pressure_vals = [s["value"] for s in sensor_data if s.get("type") == "pressure"]
        flow_vals = [s["value"] for s in sensor_data if s.get("type") == "flow_rate"]

        features = np.array([
            np.mean(temp_vals) if temp_vals else 1200,
            np.std(temp_vals) if temp_vals else 50,
            np.max(temp_vals) if temp_vals else 1500,
            np.min(temp_vals) if temp_vals else 900,
            np.mean(pressure_vals) if pressure_vals else 1.5,
            np.std(pressure_vals) if pressure_vals else 0.1,
            np.mean(flow_vals) if flow_vals else 2.5,
            np.std(flow_vals) if flow_vals else 0.2,
            len(temp_vals),
            len(pressure_vals),
        ], dtype=np.float32)

        input_tensor = torch.tensor(features).unsqueeze(0).to(device)

        with torch.no_grad():
            output = efficiency_model(input_tensor)

        vals = output.squeeze().cpu().numpy()
        result = {
            "combustion_efficiency": float(np.clip(vals[0], 0, 1)),
            "thermal_efficiency": float(np.clip(vals[1], 0, 1)),
            "heat_release_rate": float(max(0, vals[2])),
            "fuel_air_ratio": float(np.clip(vals[3], 0.01, 0.1)),
        }
        return jsonify(result)

    except Exception as e:
        logger.error(f"Efficiency prediction error: {e}")
        return jsonify({"error": str(e)}), 500


@app.route("/api/health", methods=["GET"])
def health():
    return jsonify({
        "status": "healthy",
        "device": str(device),
        "models_loaded": {
            "temperature": os.path.exists(temp_model_path),
            "instability": os.path.exists(instability_model_path),
        }
    })


def extract_features(inputs, predictions):
    pred = predictions.squeeze(1)
    max_temp = pred.amax(dim=(1, 2))
    min_temp = pred.amin(dim=(1, 2))
    mean_temp = pred.mean(dim=(1, 2))
    std_temp = pred.std(dim=(1, 2))
    grad_x = torch.diff(pred, dim=1)
    grad_y = torch.diff(pred, dim=2)
    grad_mag = torch.sqrt(grad_x[:, :, :-1] ** 2 + grad_y[:, :-1, :] ** 2)
    mean_grad = grad_mag.mean(dim=(1, 2))
    max_grad = grad_mag.amax(dim=(1, 2))
    features = torch.stack([max_temp, min_temp, mean_temp, std_temp, mean_grad, max_grad,
                            std_temp / (mean_temp + 1e-6), max_grad / (mean_temp + 1e-6)], dim=1)
    return features


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5000, debug=False)

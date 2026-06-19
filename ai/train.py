import torch
import torch.nn as nn
import torch.optim as optim
from torch.utils.data import DataLoader, TensorDataset
import numpy as np
import os
import sys

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
from models.cnn_pde_model import CNNPDEModel, InstabilityDetector, EfficiencyPredictor
from data.generator import generate_training_data


def train_temperature_model(epochs=50, batch_size=32, lr=1e-3, num_samples=2000):
    print("Generating training data...")
    inputs, targets, labels = generate_training_data(num_samples)

    split = int(0.8 * num_samples)
    train_inputs, val_inputs = inputs[:split], inputs[split:]
    train_targets, val_targets = targets[:split], targets[split:]
    train_labels, val_labels = labels[:split], labels[split:]

    train_dataset = TensorDataset(train_inputs, train_targets)
    val_dataset = TensorDataset(val_inputs, val_targets)
    train_loader = DataLoader(train_dataset, batch_size=batch_size, shuffle=True)
    val_loader = DataLoader(val_dataset, batch_size=batch_size)

    device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
    print(f"Training on device: {device}")

    model = CNNPDEModel(in_channels=4, grid_size=16).to(device)
    optimizer = optim.Adam(model.parameters(), lr=lr, weight_decay=1e-5)
    scheduler = optim.lr_scheduler.StepLR(optimizer, step_size=20, gamma=0.5)
    criterion = nn.MSELoss()

    best_val_loss = float("inf")
    model_dir = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "models")
    os.makedirs(model_dir, exist_ok=True)

    for epoch in range(epochs):
        model.train()
        train_loss = 0.0
        for batch_inputs, batch_targets in train_loader:
            batch_inputs, batch_targets = batch_inputs.to(device), batch_targets.to(device)
            optimizer.zero_grad()
            outputs = model(batch_inputs)
            loss = criterion(outputs, batch_targets)
            loss.backward()
            optimizer.step()
            train_loss += loss.item()

        scheduler.step()
        train_loss /= len(train_loader)

        model.eval()
        val_loss = 0.0
        with torch.no_grad():
            for batch_inputs, batch_targets in val_loader:
                batch_inputs, batch_targets = batch_inputs.to(device), batch_targets.to(device)
                outputs = model(batch_inputs)
                loss = criterion(outputs, batch_targets)
                val_loss += loss.item()
        val_loss /= len(val_loader)

        if val_loss < best_val_loss:
            best_val_loss = val_loss
            torch.save(model.state_dict(), os.path.join(model_dir, "temperature_cnn_pde.pt"))
            print(f"  -> Saved best model (val_loss: {val_loss:.4f})")

        if (epoch + 1) % 5 == 0:
            print(f"Epoch [{epoch+1}/{epochs}] Train Loss: {train_loss:.4f} Val Loss: {val_loss:.4f}")

    print(f"Training complete. Best val loss: {best_val_loss:.4f}")

    print("\nTraining instability detector...")
    train_instability_detector(train_inputs, train_labels, val_inputs, val_labels, device, model_dir)

    return model


def train_instability_detector(train_inputs, train_labels, val_inputs, val_labels, device, model_dir):
    temp_model = CNNPDEModel(in_channels=4, grid_size=16).to(device)
    temp_model.eval()

    with torch.no_grad():
        train_features = []
        for i in range(0, len(train_inputs), 32):
            batch = train_inputs[i:i+32].to(device)
            out = temp_model(batch)
            feat = extract_instability_features(batch, out)
            train_features.append(feat.cpu())
        train_features = torch.cat(train_features, dim=0)

        val_features = []
        for i in range(0, len(val_inputs), 32):
            batch = val_inputs[i:i+32].to(device)
            out = temp_model(batch)
            feat = extract_instability_features(batch, out)
            val_features.append(feat.cpu())
        val_features = torch.cat(val_features, dim=0)

    train_feat_seq = train_features.unsqueeze(1).repeat(1, 10, 1)
    val_feat_seq = val_features.unsqueeze(1).repeat(1, 10, 1)

    detector = InstabilityDetector(input_size=8, hidden_size=64, num_layers=2).to(device)
    optimizer = optim.Adam(detector.parameters(), lr=1e-3)
    criterion = nn.BCELoss()

    for epoch in range(20):
        detector.train()
        optimizer.zero_grad()
        outputs = detector(train_feat_seq.to(device)).squeeze()
        loss = criterion(outputs, train_labels.to(device))
        loss.backward()
        optimizer.step()

        if (epoch + 1) % 5 == 0:
            print(f"  Instability Detector Epoch [{epoch+1}/20] Loss: {loss.item():.4f}")

    torch.save(detector.state_dict(), os.path.join(model_dir, "instability_detector.pt"))
    print("  Instability detector saved.")


def extract_instability_features(inputs, predictions):
    pred = predictions.squeeze(1)
    batch_size = pred.shape[0]

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
    train_temperature_model(epochs=50, num_samples=2000)

import torch
import torch.nn as nn
import torch.nn.functional as F


class TemperatureCNN(nn.Module):
    def __init__(self, in_channels=4, grid_size=16):
        super(TemperatureCNN, self).__init__()
        self.grid_size = grid_size

        self.encoder = nn.Sequential(
            nn.Conv2d(in_channels, 64, kernel_size=3, padding=1),
            nn.BatchNorm2d(64),
            nn.ReLU(inplace=True),
            nn.Conv2d(64, 128, kernel_size=3, padding=1),
            nn.BatchNorm2d(128),
            nn.ReLU(inplace=True),
            nn.MaxPool2d(2),
            nn.Conv2d(128, 256, kernel_size=3, padding=1),
            nn.BatchNorm2d(256),
            nn.ReLU(inplace=True),
            nn.Conv2d(256, 256, kernel_size=3, padding=1),
            nn.BatchNorm2d(256),
            nn.ReLU(inplace=True),
        )

        self.attention = nn.Sequential(
            nn.Conv2d(256, 1, kernel_size=1),
            nn.Sigmoid(),
        )

        self.decoder = nn.Sequential(
            nn.ConvTranspose2d(256, 128, kernel_size=2, stride=2),
            nn.BatchNorm2d(128),
            nn.ReLU(inplace=True),
            nn.Conv2d(128, 64, kernel_size=3, padding=1),
            nn.BatchNorm2d(64),
            nn.ReLU(inplace=True),
            nn.Conv2d(64, 1, kernel_size=3, padding=1),
        )

    def forward(self, x):
        feat = self.encoder(x)
        attn = self.attention(feat)
        feat = feat * attn
        out = self.decoder(feat)
        return out


class PDELayer(nn.Module):
    def __init__(self, dx=1.0, dt=0.01, alpha=0.5, iterations=5):
        super(PDELayer, self).__init__()
        self.dx = dx
        self.dt = dt
        self.alpha = alpha
        self.iterations = iterations
        self.diffusivity = nn.Parameter(torch.tensor(alpha))

    def forward(self, temperature):
        T = temperature
        for _ in range(self.iterations):
            laplacian = self._laplacian(T)
            T = T + self.diffusivity * self.dt * laplacian
        return T

    def _laplacian(self, T):
        pad = F.pad(T, (1, 1, 1, 1), mode='replicate')
        lap = (
            pad[:, :, 2:, 1:-1] + pad[:, :, :-2, 1:-1] +
            pad[:, :, 1:-1, 2:] + pad[:, :, 1:-1, :-2] -
            4 * T
        ) / (self.dx ** 2)
        return lap


class CNNPDEModel(nn.Module):
    def __init__(self, in_channels=4, grid_size=16):
        super(CNNPDEModel, self).__init__()
        self.cnn = TemperatureCNN(in_channels, grid_size)
        self.pde = PDELayer()
        self.refine = nn.Sequential(
            nn.Conv2d(2, 16, kernel_size=3, padding=1),
            nn.ReLU(inplace=True),
            nn.Conv2d(16, 1, kernel_size=3, padding=1),
        )

    def forward(self, x):
        cnn_out = self.cnn(x)
        pde_out = self.pde(cnn_out)
        combined = torch.cat([cnn_out, pde_out], dim=1)
        refined = self.refine(combined)
        return refined


class InstabilityDetector(nn.Module):
    def __init__(self, input_size=8, hidden_size=64, num_layers=2):
        super(InstabilityDetector, self).__init__()
        self.lstm = nn.LSTM(input_size, hidden_size, num_layers, batch_first=True)
        self.classifier = nn.Sequential(
            nn.Linear(hidden_size, 32),
            nn.ReLU(inplace=True),
            nn.Dropout(0.3),
            nn.Linear(32, 1),
            nn.Sigmoid(),
        )

    def forward(self, x):
        lstm_out, _ = self.lstm(x)
        last_hidden = lstm_out[:, -1, :]
        instability_prob = self.classifier(last_hidden)
        return instability_prob


class EfficiencyPredictor(nn.Module):
    def __init__(self, input_dim=10):
        super(EfficiencyPredictor, self).__init__()
        self.net = nn.Sequential(
            nn.Linear(input_dim, 128),
            nn.ReLU(inplace=True),
            nn.BatchNorm1d(128),
            nn.Linear(128, 64),
            nn.ReLU(inplace=True),
            nn.BatchNorm1d(64),
            nn.Linear(64, 32),
            nn.ReLU(inplace=True),
            nn.Linear(32, 4),
        )

    def forward(self, x):
        return self.net(x)

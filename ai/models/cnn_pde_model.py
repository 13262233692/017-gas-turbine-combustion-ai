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
    def __init__(self, dx=1.0, dt=None, alpha=0.5, max_iterations=20,
                 cfl_safety=0.4, tolerance=1e-4, use_energy_conservation=True):
        super(PDELayer, self).__init__()
        self.dx = dx
        self.dt = dt
        self.max_iterations = max_iterations
        self.cfl_safety = cfl_safety
        self.tolerance = tolerance
        self.use_energy_conservation = use_energy_conservation
        self.diffusivity = nn.Parameter(torch.tensor(alpha))

    def forward(self, temperature):
        T = temperature.clone()
        batch_size, channels, h, w = T.shape

        if self.dt is None:
            dt_max = self.cfl_safety * self.dx ** 2 / (2 * torch.abs(self.diffusivity).clamp(min=1e-6))
            dt = torch.clamp(dt_max, max=0.05).item()
        else:
            dt = self.dt
            cfl = torch.abs(self.diffusivity) * dt / (self.dx ** 2)
            if cfl > 0.5:
                dt = 0.4 * self.dx ** 2 / torch.abs(self.diffusivity).clamp(min=1e-6)
                dt = dt.item()

        initial_energy = None
        if self.use_energy_conservation:
            initial_energy = T.sum(dim=(2, 3), keepdim=True).detach()

        residual = float('inf')
        for iteration in range(self.max_iterations):
            T_prev = T.clone()
            laplacian = self._laplacian(T)
            T_new = T + self.diffusivity * dt * laplacian

            if self.use_energy_conservation and initial_energy is not None:
                current_energy = T_new.sum(dim=(2, 3), keepdim=True)
                energy_ratio = initial_energy / (current_energy + 1e-8)
                T_new = T_new * energy_ratio

            T = T_new

            residual = torch.max(torch.abs(T - T_prev)).item()
            if residual < self.tolerance:
                break

        self.residual = residual
        return T

    def _laplacian(self, T):
        pad = F.pad(T, (1, 1, 1, 1), mode='reflect')
        lap = (
            pad[:, :, 2:, 1:-1] + pad[:, :, :-2, 1:-1] +
            pad[:, :, 1:-1, 2:] + pad[:, :, 1:-1, :-2] -
            4 * T
        ) / (self.dx ** 2)
        return lap


class CNNPDEModel(nn.Module):
    def __init__(self, in_channels=4, grid_size=16,
                 min_temp=600.0, max_temp=2200.0, max_gradient=100.0):
        super(CNNPDEModel, self).__init__()
        self.grid_size = grid_size
        self.min_temp = min_temp
        self.max_temp = max_temp
        self.max_gradient = max_gradient

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

        refined = self._apply_physical_constraints(refined)
        return refined

    def _apply_physical_constraints(self, T):
        T = torch.clamp(T, self.min_temp, self.max_temp)

        for _ in range(2):
            grad_x = T[:, :, :, 1:] - T[:, :, :, :-1]
            grad_y = T[:, :, 1:, :] - T[:, :, :-1, :]

            max_grad_x = torch.max(torch.abs(grad_x))
            max_grad_y = torch.max(torch.abs(grad_y))

            if max_grad_x <= self.max_gradient and max_grad_y <= self.max_gradient:
                break

            smooth = F.avg_pool2d(T, kernel_size=3, stride=1, padding=1)
            T = 0.7 * T + 0.3 * smooth
            T = torch.clamp(T, self.min_temp, self.max_temp)

        return T


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

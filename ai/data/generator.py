import numpy as np
import torch
from scipy.ndimage import gaussian_filter


class TemperatureDataGenerator:
    def __init__(self, grid_size=16, num_sensors=24):
        self.grid_size = grid_size
        self.num_sensors = num_sensors

    def generate_combustion_field(self, batch_size=1, with_anomaly=False):
        fields = []
        for _ in range(batch_size):
            center_temp = np.random.uniform(1100, 1500)
            field = np.zeros((self.grid_size, self.grid_size))

            for i in range(self.grid_size):
                for j in range(self.grid_size):
                    ri = (i - self.grid_size / 2) / (self.grid_size / 2)
                    rj = (j - self.grid_size / 2) / (self.grid_size / 2)
                    radius = np.sqrt(ri ** 2 + rj ** 2)
                    angle = np.arctan2(ri, rj)

                    temp = center_temp * (1.0 - 0.3 * radius ** 2)
                    temp += 50 * np.sin(3 * angle) * (1 - radius)
                    temp += np.random.normal(0, 15)
                    field[i, j] = temp

            if with_anomaly:
                cx, cy = np.random.randint(2, self.grid_size - 2, 2)
                field[cx - 1:cx + 2, cy - 1:cy + 2] += np.random.uniform(100, 300)

            field = gaussian_filter(field, sigma=1.0)
            fields.append(field)

        return np.array(fields, dtype=np.float32)

    def generate_sensor_readings(self, fields):
        readings = []
        for field in fields:
            sensor_vals = []
            for i in range(self.num_sensors):
                angle = 2 * np.pi * i / self.num_sensors
                r = np.random.uniform(0.3, 0.9)
                si = int((r * np.cos(angle) + 1) / 2 * (self.grid_size - 1))
                sj = int((r * np.sin(angle) + 1) / 2 * (self.grid_size - 1))
                si = np.clip(si, 0, self.grid_size - 1)
                sj = np.clip(sj, 0, self.grid_size - 1)
                sensor_vals.append(field[si, sj] + np.random.normal(0, 5))
            readings.append(sensor_vals)
        return np.array(readings, dtype=np.float32)

    def sensor_to_input_grid(self, sensor_readings):
        batch_size = sensor_readings.shape[0]
        grids = np.zeros((batch_size, 4, self.grid_size, self.grid_size), dtype=np.float32)

        for b in range(batch_size):
            temp_grid = np.full((self.grid_size, self.grid_size), 1200.0)
            pressure_grid = np.full((self.grid_size, self.grid_size), 1.5)
            flow_grid = np.full((self.grid_size, self.grid_size), 2.5)
            quality_grid = np.ones((self.grid_size, self.grid_size))

            for i, val in enumerate(sensor_readings[b]):
                angle = 2 * np.pi * i / self.num_sensors
                r = 0.6
                si = int((r * np.cos(angle) + 1) / 2 * (self.grid_size - 1))
                sj = int((r * np.sin(angle) + 1) / 2 * (self.grid_size - 1))
                si = np.clip(si, 0, self.grid_size - 1)
                sj = np.clip(sj, 0, self.grid_size - 1)

                for di in range(-2, 3):
                    for dj in range(-2, 3):
                        ni, nj = si + di, sj + dj
                        if 0 <= ni < self.grid_size and 0 <= nj < self.grid_size:
                            dist = np.sqrt(di ** 2 + dj ** 2)
                            w = np.exp(-dist ** 2 / 2.0)
                            temp_grid[ni, nj] = temp_grid[ni, nj] * (1 - w) + val * w

            temp_grid = gaussian_filter(temp_grid, sigma=2.0)
            grids[b, 0] = temp_grid
            grids[b, 1] = pressure_grid
            grids[b, 2] = flow_grid
            grids[b, 3] = quality_grid

        return grids


def generate_training_data(num_samples=1000, grid_size=16, num_sensors=24, anomaly_ratio=0.2):
    gen = TemperatureDataGenerator(grid_size, num_sensors)

    normal_count = int(num_samples * (1 - anomaly_ratio))
    anomaly_count = num_samples - normal_count

    normal_fields = gen.generate_combustion_field(normal_count, with_anomaly=False)
    anomaly_fields = gen.generate_combustion_field(anomaly_count, with_anomaly=True)

    all_fields = np.concatenate([normal_fields, anomaly_fields], axis=0)
    all_labels = np.concatenate([
        np.zeros(normal_count),
        np.ones(anomaly_count)
    ], axis=0)

    sensor_readings = gen.generate_sensor_readings(all_fields)
    input_grids = gen.sensor_to_input_grid(sensor_readings)

    indices = np.random.permutation(num_samples)
    return (
        torch.tensor(input_grids[indices]),
        torch.tensor(all_fields[indices]).unsqueeze(1),
        torch.tensor(all_labels[indices], dtype=torch.float32),
    )


if __name__ == "__main__":
    inputs, targets, labels = generate_training_data(100)
    print(f"Inputs shape: {inputs.shape}")
    print(f"Targets shape: {targets.shape}")
    print(f"Labels shape: {labels.shape}")
    print(f"Anomaly ratio: {labels.mean():.2f}")

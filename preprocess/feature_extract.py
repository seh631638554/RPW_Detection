#!/usr/bin/env python3
import argparse
import json
from pathlib import Path

import librosa
import numpy as np
import soundfile as sf
import torch

N_FFT = 2048
WIN_LENGTH = 2048
HOP_LENGTH = 1024
WINDOW = "hann"
CENTER = True
EPS = 1e-10
FMIN = 0
FMAX = 4000
MEL_POWER = 2.0


def parse_args():
    parser = argparse.ArgumentParser(description="Extract log-mel and PCEN features from wav")
    parser.add_argument("--input", required=True)
    parser.add_argument("--output", required=True)
    parser.add_argument("--feature", default="log_mel_pcen")
    parser.add_argument("--sample-rate", type=int, default=8000)
    parser.add_argument("--n-mels", type=int, default=48)
    parser.add_argument("--fixed-seconds", type=float, default=20.0)
    return parser.parse_args()


def load_audio(path: str):
    audio, sample_rate = sf.read(path)
    if audio.ndim == 2:
        audio = audio[:, 0]
    audio = np.asarray(audio, dtype=np.float32)
    return audio, sample_rate


def fix_length_center(audio: np.ndarray, target_len: int):
    n = len(audio)
    if n == target_len:
        return audio
    if n > target_len:
        start = (n - target_len) // 2
        return audio[start:start + target_len]

    out = np.zeros((target_len,), dtype=audio.dtype)
    out[:n] = audio
    return out


def build_mel_features(
    audio: np.ndarray,
    sample_rate: int,
    target_rate: int,
    n_mels: int,
    fixed_seconds: float,
):
    if sample_rate != target_rate:
        audio = librosa.resample(audio, orig_sr=sample_rate, target_sr=target_rate)
        sample_rate = target_rate

    audio = np.nan_to_num(audio, nan=0.0, posinf=0.0, neginf=0.0).astype(np.float32, copy=False)
    if fixed_seconds > 0:
        target_len = int(round(sample_rate * fixed_seconds))
        audio = fix_length_center(audio, target_len)

    mel = librosa.feature.melspectrogram(
        y=audio,
        sr=sample_rate,
        n_fft=N_FFT,
        hop_length=HOP_LENGTH,
        win_length=WIN_LENGTH,
        window=WINDOW,
        center=CENTER,
        power=MEL_POWER,
        n_mels=n_mels,
        fmin=FMIN,
        fmax=min(FMAX, sample_rate // 2),
    ).astype(np.float32)
    mel = np.nan_to_num(mel, nan=0.0, posinf=0.0, neginf=0.0)

    log_mel = np.log(mel + EPS).astype(np.float32)
    pcen = librosa.pcen(mel, sr=sample_rate, hop_length=HOP_LENGTH).astype(np.float32)
    pcen = np.nan_to_num(pcen, nan=0.0, posinf=0.0, neginf=0.0)

    return log_mel, pcen, sample_rate


def select_feature_tensor(feature_mode: str, log_mel: np.ndarray, pcen: np.ndarray):
    feature_mode = feature_mode.lower()
    if feature_mode == "log_mel":
        tensor = torch.from_numpy(np.expand_dims(log_mel, axis=0))
        feature_type = "log_mel"
        inputs = ["log_mel"]
    elif feature_mode == "pcen":
        tensor = torch.from_numpy(np.expand_dims(pcen, axis=0))
        feature_type = "pcen"
        inputs = ["pcen"]
    elif feature_mode in {"log_mel_pcen", "logmel_pcen", "dual"}:
        tensor = {
            "log_mel": torch.from_numpy(np.expand_dims(log_mel, axis=0)),
            "pcen": torch.from_numpy(np.expand_dims(pcen, axis=0)),
        }
        feature_type = "log_mel_pcen"
        inputs = ["log_mel", "pcen"]
    else:
        raise ValueError(f"unsupported feature type: {feature_mode}")
    return tensor, feature_type, inputs


def main():
    args = parse_args()
    audio, sample_rate = load_audio(args.input)
    log_mel, pcen, final_rate = build_mel_features(
        audio,
        sample_rate,
        args.sample_rate,
        args.n_mels,
        args.fixed_seconds,
    )
    feature_tensor, feature_type, inputs = select_feature_tensor(args.feature, log_mel, pcen)

    output_path = Path(args.output)
    output_path.parent.mkdir(parents=True, exist_ok=True)
    torch.save(feature_tensor, output_path)

    if isinstance(feature_tensor, dict):
        sample_tensor = feature_tensor[inputs[0]]
    else:
        sample_tensor = feature_tensor

    print(
        json.dumps(
            {
                "output_path": str(output_path),
                "feature_type": feature_type,
                "inputs": inputs,
                "shape": list(sample_tensor.shape),
                "dtype": str(sample_tensor.dtype).replace("torch.", ""),
                "sample_rate": final_rate,
                "n_mels": args.n_mels,
                "fixed_seconds": args.fixed_seconds,
            }
        )
    )


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
import argparse
import json
from pathlib import Path

import torch

from model_def import mobilenetv3_ultra_light_mc


def parse_args():
    parser = argparse.ArgumentParser(description="Run audio classification inference")
    parser.add_argument("--input", required=True)
    parser.add_argument("--model", required=True)
    parser.add_argument("--device", default="cpu")
    parser.add_argument("--labels", default="clean,infested")
    return parser.parse_args()


def ensure_input(tensor):
    if not torch.is_tensor(tensor):
        tensor = torch.as_tensor(tensor)
    tensor = tensor.detach().float()

    if tensor.ndim == 2:
        tensor = tensor.unsqueeze(0).unsqueeze(0)
    elif tensor.ndim == 3:
        tensor = tensor.unsqueeze(0)
    elif tensor.ndim != 4:
        raise ValueError(f"unsupported tensor shape: {tuple(tensor.shape)}")

    return tensor


def load_feature(path: str):
    obj = torch.load(path, map_location="cpu")
    if isinstance(obj, dict):
        if "log_mel" not in obj:
            raise ValueError("feature dict must contain log_mel")
        log_mel = obj["log_mel"]
    elif torch.is_tensor(obj):
        if obj.ndim == 2:
            log_mel = obj.unsqueeze(0)
        elif obj.ndim == 3:
            log_mel = obj[0].unsqueeze(0) if obj.shape[0] > 1 else obj
        else:
            raise ValueError(f"unsupported tensor shape: {tuple(obj.shape)}")
    else:
        raise ValueError(f"unsupported feature object type: {type(obj)}")

    return ensure_input(log_mel)


def load_model(path: str, device: str, num_classes: int):
    model_path = Path(path)
    if not model_path.exists():
        raise FileNotFoundError(f"model not found: {model_path}")

    try:
        model = torch.jit.load(str(model_path), map_location=device)
        if hasattr(model, "eval"):
            model.eval()
        return model, model_path.name
    except Exception:
        state = torch.load(str(model_path), map_location="cpu")

    if isinstance(state, dict) and "state_dict" in state:
        state = state["state_dict"]
    if not isinstance(state, dict):
        raise ValueError("unsupported model file, expected TorchScript or state_dict checkpoint")

    cleaned_state = {}
    for key, value in state.items():
        if key.startswith("module."):
            key = key[len("module."):]
        cleaned_state[key] = value

    model = mobilenetv3_ultra_light_mc(num_classes=num_classes)
    model.load_state_dict(cleaned_state, strict=True)
    model.to(device)
    model.eval()

    return model, model_path.name


def normalize_like_training(x):
    mean = x.mean()
    std = x.std()
    return (x - mean) / (std + 1e-6)


def run_model(model, log_mel):
    with torch.no_grad():
        outputs = model(normalize_like_training(log_mel))

    if isinstance(outputs, (list, tuple)):
        outputs = outputs[0]
    elif isinstance(outputs, dict):
        if "logits" in outputs:
            outputs = outputs["logits"]
        elif "output" in outputs:
            outputs = outputs["output"]
        else:
            outputs = next(iter(outputs.values()))

    if not torch.is_tensor(outputs):
        outputs = torch.as_tensor(outputs)

    outputs = outputs.detach().float().cpu().reshape(-1)

    if outputs.numel() == 1:
        infested_prob = torch.sigmoid(outputs[0]).item()
        return [1.0 - infested_prob, infested_prob]

    probs = torch.softmax(outputs, dim=0)
    return probs.tolist()


def build_probabilities(labels, probs):
    if len(labels) < len(probs):
        labels = labels + [f"class_{i}" for i in range(len(labels), len(probs))]
    elif len(labels) > len(probs):
        labels = labels[:len(probs)]
    return {label: float(prob) for label, prob in zip(labels, probs)}, labels


def main():
    args = parse_args()
    labels = [item.strip() for item in args.labels.split(",") if item.strip()]
    if not labels:
        labels = ["clean", "infested"]

    log_mel = load_feature(args.input)
    device = torch.device(args.device)
    model, model_name = load_model(args.model, device, len(labels))
    probs = run_model(model, log_mel.to(device))
    probabilities, labels = build_probabilities(labels, probs)

    pred_index = int(max(range(len(probs)), key=lambda i: probs[i]))
    pred_label = labels[pred_index]
    score = float(probs[pred_index])

    print(
        json.dumps(
            {
                "label": pred_label,
                "index": pred_index,
                "score": score,
                "probabilities": probabilities,
                "model_name": model_name,
            }
        )
    )


if __name__ == "__main__":
    main()

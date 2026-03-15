#!/usr/bin/env python3
from typing import List

import torch
from torch import nn
import torch.nn.functional as F


class SEModule(nn.Module):
    def __init__(self, channels, reduction=4):
        super().__init__()
        self.fc1 = nn.Linear(channels, channels // reduction)
        self.fc2 = nn.Linear(channels // reduction, channels)
        self.sigmoid = nn.Sigmoid()

    def forward(self, x):
        y = F.adaptive_avg_pool2d(x, 1)
        y = y.view(y.size(0), -1)
        y = self.fc1(y)
        y = F.relu(y, inplace=True)
        y = self.fc2(y)
        y = self.sigmoid(y).unsqueeze(-1).unsqueeze(-1)
        return x * y


class ECAModule(nn.Module):
    def __init__(self, channels, kernel_size=3):
        super().__init__()
        self.avg_pool = nn.AdaptiveAvgPool2d(1)
        self.conv1d = nn.Conv1d(
            in_channels=channels,
            out_channels=channels,
            kernel_size=kernel_size,
            padding=(kernel_size // 2),
            groups=channels,
            bias=False,
        )
        self.sigmoid = nn.Sigmoid()

    def forward(self, x):
        y = self.avg_pool(x)
        y = y.view(y.size(0), y.size(1), -1)
        y = self.conv1d(y)
        y = y.view(y.size(0), y.size(1), 1, 1)
        y = self.sigmoid(y)
        return x * y


class MultiSpectralChannelAttention(nn.Module):
    def __init__(self, channels, freq_sel=None):
        super().__init__()
        if freq_sel is None:
            freq_sel = [1, 2, 4, 8, 16, 32]
        self.fc = nn.ModuleList([nn.Linear(channels, channels) for _ in freq_sel])
        self.combine = nn.Linear(channels * len(freq_sel), channels)
        self.sigmoid = nn.Sigmoid()

    def forward(self, x):
        y = torch.mean(x, dim=(2, 3))
        attn = torch.cat([fc(y) for fc in self.fc], dim=1)
        attn = self.combine(attn)
        attn = self.sigmoid(attn).unsqueeze(-1).unsqueeze(-1)
        return x * attn


class MixConv(nn.Module):
    def __init__(self, in_channels, kernel_sizes, stride=1):
        super().__init__()
        self.splits = self._split_channels(in_channels, len(kernel_sizes))
        self.convs = nn.ModuleList()
        self.bns = nn.ModuleList()
        self.acts = nn.ModuleList()

        for c, k in zip(self.splits, kernel_sizes):
            self.convs.append(
                nn.Conv2d(
                    in_channels=c,
                    out_channels=c,
                    kernel_size=k,
                    stride=stride,
                    padding=k // 2,
                    groups=c,
                    bias=False,
                )
            )
            self.bns.append(nn.BatchNorm2d(c))
            self.acts.append(nn.ReLU(inplace=True))

    @staticmethod
    def _split_channels(channels, num_splits):
        base = channels // num_splits
        rem = channels % num_splits
        return [base + (1 if i < rem else 0) for i in range(num_splits)]

    def forward(self, x):
        x_slices = x.split(self.splits, dim=1)
        y = [
            act(bn(conv(xi)))
            for conv, bn, act, xi in zip(self.convs, self.bns, self.acts, x_slices)
        ]
        return torch.cat(y, dim=1)


class InvertedResidualConfig:
    def __init__(
        self,
        input_c: int,
        expanded_c: int,
        out_c: int,
        activation: str,
        stride: int,
        expand_ratio: int,
        use_attention: str = None,
        use_mixconv: bool = False,
        kernel_sizes: List[int] = None,
    ):
        self.input_c = input_c
        self.expanded_c = expanded_c
        self.out_c = out_c
        self.activation = activation
        self.stride = stride
        self.expand_ratio = expand_ratio
        self.use_attention = use_attention
        self.use_mixconv = use_mixconv
        self.kernel_sizes = kernel_sizes if kernel_sizes is not None else [3]


class InvertedResidual(nn.Module):
    def __init__(self, cnf: InvertedResidualConfig):
        super().__init__()
        layers: List[nn.Module] = []
        mid_c = cnf.input_c * cnf.expand_ratio
        act = nn.ReLU if cnf.activation == "RE" else nn.Hardswish

        if cnf.expand_ratio != 1:
            layers += [
                nn.Conv2d(cnf.input_c, mid_c, 1, 1, 0, bias=False),
                nn.BatchNorm2d(mid_c),
                act(inplace=True),
            ]
        else:
            mid_c = cnf.input_c

        if cnf.use_mixconv:
            layers.append(MixConv(mid_c, cnf.kernel_sizes, stride=cnf.stride))
        else:
            layers += [
                nn.Conv2d(mid_c, mid_c, 3, cnf.stride, 1, groups=mid_c, bias=False),
                nn.BatchNorm2d(mid_c),
                act(inplace=True),
            ]

        if cnf.use_attention == "MUL":
            layers.append(MultiSpectralChannelAttention(mid_c))
        elif cnf.use_attention == "ECA":
            layers.append(ECAModule(mid_c))
        elif cnf.use_attention == "SE":
            layers.append(SEModule(mid_c))

        layers += [
            nn.Conv2d(mid_c, cnf.out_c, 1, 1, 0, bias=False),
            nn.BatchNorm2d(cnf.out_c),
        ]

        self.block = nn.Sequential(*layers)
        self.use_res_connect = cnf.stride == 1 and cnf.input_c == cnf.out_c

    def forward(self, x):
        out = self.block(x)
        if self.use_res_connect:
            return x + out
        return out


class mobilenetv3_ultra_light_mc(nn.Module):
    def __init__(self, num_classes=2):
        super().__init__()
        self.conv1 = nn.Sequential(
            nn.Conv2d(1, 4, kernel_size=7, stride=2, padding=3, bias=False),
            nn.BatchNorm2d(4),
            nn.ReLU(inplace=True),
        )
        self.bneck1 = InvertedResidual(
            InvertedResidualConfig(4, 4, 8, "RE", 1, expand_ratio=1, use_attention="NONE", use_mixconv=False)
        )
        self.bneck2 = InvertedResidual(
            InvertedResidualConfig(8, 24, 16, "RE", 2, expand_ratio=3, use_attention="ECA", use_mixconv=False)
        )
        self.bneck3 = InvertedResidual(
            InvertedResidualConfig(16, 48, 32, "RE", 1, expand_ratio=3, use_attention="ECA", use_mixconv=True, kernel_sizes=[5, 7])
        )
        self.bneck4 = InvertedResidual(
            InvertedResidualConfig(32, 32, 64, "HS", 2, expand_ratio=1, use_attention="ECA", use_mixconv=True, kernel_sizes=[5, 7])
        )
        self.avgpool = nn.AdaptiveAvgPool2d((1, 1))
        self.classifier = nn.Linear(64, num_classes)

    def forward(self, x):
        x = self.conv1(x)
        x = self.bneck1(x)
        x = self.bneck2(x)
        x = self.bneck3(x)
        x = self.bneck4(x)
        x = self.avgpool(x)
        x = x.view(x.size(0), -1)
        x = self.classifier(x)
        return x

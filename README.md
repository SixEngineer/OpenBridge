# OpenBridge

> ⚡ OpenBridge — Complementary bridge service for OpenList
> OpenBridge 是 OpenList 的配套桥接服务，用于补全 OpenList 部分存储驱动缺失的**总容量、已用空间、剩余容量查询能力**，对外提供标准化容量接口，兼容 OpenList API 规范，轻量易部署。

[![GitHub Stars](https://img.shields.io/github/stars/SixEngineer/OpenBridge)](https://github.com/SixEngineer/OpenBridge)
[![GitHub Forks](https://img.shields.io/github/forks/SixEngineer/OpenBridge)](https://github.com/SixEngineer/OpenBridge)
[![License](https://img.shields.io/github/license/SixEngineer/OpenBridge)](LICENSE)

**OpenBridge** works as an auxiliary service for OpenList.
Some storage drivers in OpenList lack the capacity‑query capability, cannot return total/used/free storage space.
OpenBridge implements the missing capacity interface, returning standardized storage capacity data to OpenList.

## ✨ Features
- 📦 补全 OpenList 缺失的存储容量查询（总容量 / 已用 / 剩余空间）
- 🔌 兼容 OpenList API 格式，可直接对接
- ⚡ 轻量高性能，低资源开销
- 🖥️ 跨平台部署，简单配置即可运行

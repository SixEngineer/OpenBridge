# OpenBridge

基于 OpenList 的直链下载编排与容量适配系统

## 技术栈
- Go (Backend)
- Vue3 (Frontend)
- aria2 (Download Engine)

## WebDAV 挂载

如果你希望：

1. 文件操作继续复用 OpenList
2. 本地盘符容量显示改为 OpenBridge `mount` 层容量

可以使用新增的按挂载点代理的 WebDAV 地址：

```text
/api/v1/webdav/mounts/:id
```

完整使用说明见：

- [docs/webdav_mount.md](docs/webdav_mount.md)

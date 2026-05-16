# GPT88 在 1Panel 中部署

本仓库提供了适合 1Panel 的源码构建版 Compose 文件：

- `deploy/docker-compose.1panel.yml`
- `deploy/.env.1panel.example`

## 部署步骤

1. 在服务器上克隆仓库

```bash
git clone https://github.com/Osub/sub2api.git
cd sub2api/deploy
```

2. 复制环境文件并修改关键变量

```bash
cp .env.1panel.example .env
```

至少要修改：

- `POSTGRES_PASSWORD`
- `JWT_SECRET`
- `ADMIN_EMAIL`
- `ADMIN_PASSWORD`（可留空自动生成）

如果服务器访问 npm 官方源不稳定，保留默认值即可：

- `PNPM_VERSION=10.20.0`
- `NPM_REGISTRY=https://registry.npmmirror.com`

3. 在 1Panel 中创建 Compose 应用

- 编排文件选择 `deploy/docker-compose.1panel.yml`
- 工作目录使用仓库中的 `deploy/`
- 首次启动选择构建并启动

4. 启动后访问

```text
http://服务器IP:8080
```

如果你在 `.env` 中修改了 `SERVER_PORT`，请使用对应端口。

## 更新流程

服务器上执行：

```bash
cd /path/to/sub2api
git pull
```

然后在 1Panel 中对该 Compose 应用执行：

- 重新构建
- 重启

这样可以把 GitHub 上最新的二开代码重新构建并部署到线上。

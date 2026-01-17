
# Assistant

## 配置

### 配置文件

项目使用 `config.yaml` 配置文件。首次运行前，请复制示例配置：

```bash
cp config.example.yaml config.yaml
```

然后根据需要修改 `config.yaml` 中的配置项。

### Admin 用户配置

系统会在启动时自动创建管理员用户。可以在 `config.yaml` 中配置：

```yaml
app:
  root:
    username: admin
    password: your_secure_password
    email: admin@example.com
```

### 分布式锁配置

系统集成了基于数据库的分布式锁功能，用于多实例间的资源协调：

```yaml
dislock:
  default_ttl: 30    # 默认锁TTL（秒）
  max_ttl: 300       # 最大锁TTL（秒）
```

详细使用说明请参考 [docs/distributed-lock.md](docs/distributed-lock.md)。

详细配置说明请参考 [docs/admin-config.md](docs/admin-config.md)。

### 启动

```bash
make run
```

### 部署

```bash
docker compose up --build -d
```

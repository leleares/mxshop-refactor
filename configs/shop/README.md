# 配置文件说明

## 使用方法

1. 复制 `api.yaml.example` 为 `api.yaml`：
   ```bash
   cp api.yaml.example api.yaml
   ```

2. 编辑 `api.yaml`，填入真实的配置信息：
   - `sms.key`: 你的阿里云 AccessKey ID
   - `sms.secret`: 你的阿里云 AccessKey Secret
   - 其他需要自定义的配置项

3. `api.yaml` 文件已被添加到 `.gitignore`，不会被提交到 git 仓库

## 注意事项

- **绝对不要**将包含真实密钥的 `api.yaml` 提交到代码仓库
- 如果密钥不慎泄露，请立即到阿里云控制台轮换密钥
- 生产环境建议使用环境变量或密钥管理服务

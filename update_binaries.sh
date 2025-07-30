#!/bin/bash

# CMDB二进制文件更新脚本
# 备份旧的二进制文件并复制新的二进制文件到data/cmdb目录

cd /Users/kiyoliang/workspacefc/bk-cmdb/src/web_server && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make DISABLE_CRYPTO=true TARGET_NAME=cmdb_webserver
cd /Users/kiyoliang/workspacefc/bk-cmdb/src/apiserver  && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make DISABLE_CRYPTO=true TARGET_NAME=cmdb_apiserver
cd /Users/kiyoliang/workspacefc/bk-cmdb/src/scene_server/admin_server && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make DISABLE_CRYPTO=true TARGET_NAME=cmdb_adminserver
cd /Users/kiyoliang/workspacefc/bk-cmdb/src/source_controller/coreservice && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make DISABLE_CRYPTO=true TARGET_NAME=cmdb_coreservice
cd /Users/kiyoliang/workspacefc/bk-cmdb/src/scene_server/topo_server && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make DISABLE_CRYPTO=true TARGET_NAME=cmdb_toposerver

set -e

# 定义路径
BUILD_DIR="/Users/kiyoliang/workspacefc/bk-cmdb/src/bin/build/master"
DATA_DIR="/Users/kiyoliang/workspacefc/bk-cmdb/data/cmdb"
BACKUP_DIR="/Users/kiyoliang/workspacefc/bk-cmdb/data/cmdb_backup_$(date +%Y%m%d_%H%M%S)"

# 要更新的服务列表
SERVICES=(
    "cmdb_webserver"
    "cmdb_apiserver"
    "cmdb_adminserver"
    "cmdb_coreservice"
    "cmdb_toposerver"
)

echo "开始更新CMDB二进制文件..."
echo "构建目录: $BUILD_DIR"
echo "数据目录: $DATA_DIR"
echo "备份目录: $BACKUP_DIR"

# 创建备份目录
echo "创建备份目录: $BACKUP_DIR"
mkdir -p "$BACKUP_DIR"

# 遍历服务列表进行备份和更新
for service in "${SERVICES[@]}"; do
    echo "处理服务: $service"
    
    # 检查源文件是否存在
    if [ ! -f "$BUILD_DIR/$service/$service" ]; then
        echo "警告: 源文件不存在 $BUILD_DIR/$service/$service"
        continue
    fi
    
    # 检查目标目录是否存在
    if [ ! -d "$DATA_DIR/$service" ]; then
        echo "警告: 目标目录不存在 $DATA_DIR/$service"
        continue
    fi
    
    # 备份旧的二进制文件（只备份二进制文件，不是整个目录）
    if [ -f "$DATA_DIR/$service/$service" ]; then
        echo "  备份旧二进制文件: $service"
        cp "$DATA_DIR/$service/$service" "$BACKUP_DIR/${service}_$(date +%Y%m%d_%H%M%S)"
    else
        echo "  旧二进制文件不存在，跳过备份: $service"
    fi
    
    # 直接覆盖复制新的二进制文件
    echo "  覆盖复制新二进制文件: $service"
    cp "$BUILD_DIR/$service/$service" "$DATA_DIR/$service/$service"
    
    # 设置执行权限
    chmod +x "$DATA_DIR/$service/$service"
    
    echo "  ✓ $service 更新完成"
done

echo ""
echo "所有二进制文件更新完成！"
echo "备份文件保存在: $BACKUP_DIR"
echo ""
echo "可以使用以下命令验证文件:"
for service in "${SERVICES[@]}"; do
    echo "ls -la $DATA_DIR/$service/$service"
done
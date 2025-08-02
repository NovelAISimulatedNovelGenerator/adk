#!/bin/sh

# 插件自动编译脚本
# 自动扫描flows目录下的所有插件并编译成.so文件

set -e  # 遇到错误立即退出

FLOWS_DIR="./flows"
# 检测运行环境，Docker中使用绝对路径，本地使用相对路径
if [ -w "/app" ]; then
    PLUGINS_DIR="/app/plugins"
else
    PLUGINS_DIR="./plugins"
fi
BUILD_ERRORS=""
ERROR_COUNT=0

# 创建插件输出目录
mkdir -p "$PLUGINS_DIR"

echo "开始扫描并编译flows目录下的插件..."
echo "flows目录: $FLOWS_DIR"
echo "插件输出目录: $PLUGINS_DIR"
echo "==========================================="

# 遍历flows目录下的所有子目录
for plugin_dir in "$FLOWS_DIR"/*/ ; do
    if [ -d "$plugin_dir" ]; then
        plugin_name=$(basename "$plugin_dir")
        main_file="$plugin_dir/main.go"
        
        # 检查是否存在main.go文件
        if [ -f "$main_file" ]; then
            so_file="$PLUGINS_DIR/${plugin_name}.so"
            
            echo "编译插件: $plugin_name"
            echo "  源文件: $main_file"
            echo "  目标文件: $so_file"
            
            # 使用CGO编译插件
            if CGO_ENABLED=1 GOOS=linux go build -buildmode=plugin -o "$so_file" "$main_file"; then
                echo "  ✅ 编译成功: $plugin_name"
            else
                echo "  ❌ 编译失败: $plugin_name"
                if [ -z "$BUILD_ERRORS" ]; then
                    BUILD_ERRORS="$plugin_name"
                else
                    BUILD_ERRORS="$BUILD_ERRORS $plugin_name"
                fi
                ERROR_COUNT=$((ERROR_COUNT + 1))
            fi
            echo ""
        else
            echo "⚠️  跳过 $plugin_name (没有找到main.go文件)"
        fi
    fi
done

echo "==========================================="
echo "插件编译完成！"

# 统计编译结果
total_plugins=$(find "$FLOWS_DIR" -name "main.go" | wc -l)
successful_plugins=$((total_plugins - ERROR_COUNT))

echo "总插件数: $total_plugins"
echo "编译成功: $successful_plugins"
echo "编译失败: $ERROR_COUNT"

# 如果有编译失败的插件，显示详情
if [ $ERROR_COUNT -gt 0 ]; then
    echo ""
    echo "编译失败的插件:"
    for error_plugin in $BUILD_ERRORS; do
        echo "  - $error_plugin"
    done
    echo ""
    echo "⚠️  注意: 有插件编译失败，请检查代码或依赖"
fi

# 显示编译结果
echo ""
echo "已编译的插件文件:"
ls -la "$PLUGINS_DIR"/*.so 2>/dev/null || echo "  (没有找到已编译的.so文件)"

echo ""
echo "插件编译脚本执行完成。"

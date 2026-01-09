#!/bin/bash
echo "=== SSH Manager CLI 功能测试 ==="
echo ""

# 清理旧配置
rm -f ~/.ssh_manager_config.yaml

echo "1. 测试初始化 (init)"
echo -e "mypassword123\nmypassword123" | /Users/qiubowen/tools/ssh-manager-go/sshmgr init
if [ $? -eq 0 ]; then
    echo "✅ 初始化成功"
else
    echo "❌ 初始化失败"
fi
echo ""

echo "2. 查看生成的YAML配置文件"
cat ~/.ssh_manager_config.yaml
echo ""

echo "3. 测试添加服务器 (add)"
echo -e "mypassword123\ntestserver\n192.168.1.100\nroot\nmypassword123\n22\nn" | /Users/qiubowen/tools/ssh-manager-go/sshmgr add
if [ $? -eq 0 ]; then
    echo "✅ 添加成功"
else
    echo "❌ 添加失败"
fi
echo ""

echo "4. 查看YAML配置（添加后）"
cat ~/.ssh_manager_config.yaml
echo ""

echo "5. 测试列出所有服务器 (list)"
echo "mypassword123" | /Users/qiubowen/tools/ssh-manager-go/sshmgr list
if [ $? -eq 0 ]; then
    echo "✅ 列出成功"
else
    echo "❌ 列出失败"
fi
echo ""

echo "6. 测试修改服务器 (modify)"
echo -e "mypassword123\n\n\n\n2222\nn" | /Users/qiubowen/tools/ssh-manager-go/sshmgr modify testserver
if [ $? -eq 0 ]; then
    echo "✅ 修改成功"
else
    echo "❌ 修改失败"
fi
echo ""

echo "7. 查看YAML配置（修改后）"
cat ~/.ssh_manager_config.yaml
echo ""

echo "8. 测试删除服务器 (delete)"
echo -e "mypassword123\ny" | /Users/qiubowen/tools/ssh-manager-go/sshmgr delete testserver
if [ $? -eq 0 ]; then
    echo "✅ 删除成功"
else
    echo "❌ 删除失败"
fi
echo ""

echo "9. 查看YAML配置（删除后）"
cat ~/.ssh_manager_config.yaml
echo ""

echo "=== 所有测试完成 ==="

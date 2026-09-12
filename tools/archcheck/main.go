package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	root, _ := os.Getwd()
	errors := 0
	appBytes, err := os.ReadFile(filepath.Join(root, "app.go"))
	if err != nil {
		fmt.Println("cannot read app.go:", err)
		os.Exit(1)
	}
	app := string(appBytes)
	if strings.Contains(app, "switch toolName") {
		fmt.Println("❌ [R2] app.go 使用 switch toolName hardcode 工具路由")
		errors++
	}

	engineDir := filepath.Join(root, "internal", "core", "loop")
	engineOK := false
	_ = filepath.Walk(engineDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		b, _ := os.ReadFile(path)
		s := string(b)
		if strings.Contains(s, "OnBeforeAct") && strings.Contains(s, "OnAfterAct") {
			engineOK = true
		}
		if strings.Contains(s, `"tiancode/plugins/tool/`) || strings.Contains(s, `"tcode/plugins/tool/`) {
			fmt.Println("❌ [R4] core/loop 直接 import 具体 tool 插件")
			errors++
		}
		return nil
	})
	if !engineOK {
		fmt.Println("❌ [R6] ExecutionEngine 缺少 Rail.OnBeforeAct / OnAfterAct")
		errors++
	}

	_ = filepath.Walk(filepath.Join(root, "plugins"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		b, _ := os.ReadFile(path)
		s := string(b)
		if strings.Contains(s, `"tiancode/app"`) || strings.Contains(s, `"tcode/app"`) {
			fmt.Println("❌ [R5] plugin 依赖宿主:", path)
			errors++
		}
		return nil
	})

	if errors > 0 {
		fmt.Printf("❌ 架构守卫失败（%d 项）\n", errors)
		os.Exit(1)
	}
	fmt.Println("✅ 架构守卫通过（热插拔插件架构合规）")
}

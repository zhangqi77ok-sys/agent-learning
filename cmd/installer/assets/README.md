# 安装包内嵌资产

打安装包前把编译产物拷到本目录：

```
Copy-Item bin/tiancode.exe cmd/installer/assets/tiancode.exe -Force
Copy-Item bin/uninstall.exe cmd/installer/assets/uninstall.exe -Force
```

`tiancode.exe` / `uninstall.exe` 不进 git。

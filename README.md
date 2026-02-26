# ecwatch

Windows 平台下直接读取 `\\.\ACPIDriver` 的 EC 寄存器，输出 CPU/GPU 风扇数据，并可选写入 HWiNFO 自定义传感器（仅 `Fan0`/`Fan1`）。

适用于清华同方主板（Tongfang）机型，已在机械革命（Mechrevo）16 上测试成功。

## 功能

- 每秒读取风扇数据（默认 1s）：
  - CPU RPM: `0x0464/0x0465`
  - GPU RPM: `0x046C/0x046B`
  - CPU Duty: `0x075B`
  - GPU Duty: `0x075C`
  - Fan Alert: `0x0741`
- 可选写入 HWiNFO：
  - `HKCU\Software\HWiNFO64\Sensors\Custom\<Group>\Fan0`
  - `HKCU\Software\HWiNFO64\Sensors\Custom\<Group>\Fan1`

## 前置条件

- Windows
- 管理员权限（需要打开 `\\.\ACPIDriver`）
- 目标机器已加载对应 ACPI 驱动（否则会报 `The system cannot find the file specified`）
- 如使用 `-hwinfo`，请确保进程运行用户与 HWiNFO 使用用户一致（同一 `HKCU`）

## 构建

在 `ecwatch` 目录执行：

```powershell
go mod tidy
go test ./...
go build -o ecwatch.exe
```

交叉编译 Windows：

```powershell
GOOS=windows GOARCH=amd64 go build -o ecwatch_windows_amd64.exe
```

## 运行

最常用：

```powershell
.\ecwatch.exe -interval 1s -debug
```

启用 HWiNFO 输出：

```powershell
.\ecwatch.exe -interval 1s -hwinfo -hwinfo-group ECWatch -debug
```

只读取一次：

```powershell
.\ecwatch.exe -once -debug
```

## 参数

- `-device`：ACPI 设备路径，默认 `\\.\ACPIDriver`
- `-interval`：轮询间隔，默认 `1s`
- `-once`：只读一次后退出
- `-debug`：输出寄存器级调试日志
- `-hwinfo`：将 CPU/GPU 风扇 RPM 写入 HWiNFO 自定义传感器
- `-hwinfo-group`：HWiNFO 分组名，默认 `ECWatch`

## 常见问题

1. `open \\.\ACPIDriver failed: The system cannot find the file specified.`

- 含义：驱动设备对象不存在。
- 处理：确认驱动服务已安装并启动，而不是权限问题。

2. 服务启动后 HWiNFO 看不到数据

- 常见原因：服务账号与 HWiNFO 登录用户不同，写入了另一个用户的 `HKCU`。
- 建议：使用同一用户运行，或改成登录用户上下文启动。

基于 Go 实现的 DrainNet 项目，一款后端服务，完成城市排水泵站降雨峰值预降、主备泵切换、闸门顺序启闭与渠道倒灌判定的联合调度。

DrainNet 提供汇水分区、泵站主备泵、闸门、雨量计、液位计、调度策略、排水配额与审计记录的 JSON API 控制台，全部数据以文件方式持久化，无需数据库。

## 构建

```bash
docker build -f benzhi.Dockerfile -t drainnet .
```

支持 amd64 与 arm64：

```bash
./build_benzhi_docker.sh drainnet linux/amd64
./build_benzhi_docker.sh drainnet linux/arm64
```

## 运行

```bash
docker run --rm -p 8080:8080 -e DRAINNET_ADDR=0.0.0.0:8080 -e DRAINNET_DATA=/app/data drainnet
```

容器内验证：

```bash
go build ./...
go test ./...
go vet ./...
```

## 接口探测

```bash
curl -s http://127.0.0.1:8080/api/v1/health
curl -s http://127.0.0.1:8080/api/v1/status
```

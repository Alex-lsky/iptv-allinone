# **使用说明：**
**如果你没有软路由或者服务器，那么推荐白嫖Vercel使用，[点击查看部署方法](https://github.com/papagaye744/iptv-go)！**

## 配置

本项目使用 `config.json` 文件进行配置。默认配置文件如下：

```json
{
  "server": {
    "port": ":35455"
  },
  "security": {
    "aes_key": "6354127897263145",
    "default_ad_url_base64": "Dy0RPTwkLOSAi3QwoeiO5LCMnrV5rKJVH/en6xEmxVk="
  },
  "urls": {
    "default_live_prefix": "https://www.goodiptv.club",
    "huya_api_base": "https://live.cdn.huya.com/liveHttpUI/getLiveList",
    "douyu_api_base": "https://www.douyu.com/gapi/rkc/directory/mixList/2_208",
    "yy_api_base": "https://rubiks-idx.yy.com/nav/other/pnk1/448772",
    "iptv_js_list_url": "http://live.epg.gitv.tv/tagNewestEpgList/JS_CUCC/1/100/0.json"
  },
  "defaults": {
    "huya_gid": "2135",
    "douyu_gid": "2_208",
    "stream_type": "flv",
    "huya_cdn": "hwcdn",
    "huya_media": "flv",
    "huya_response_type": "nodisplay",
    "bilibili_platform": "web",
    "bilibili_quality": "10000",
    "bilibili_line": "first",
    "youtube_quality": "1080",
    "yy_quality": "4"
  },
  "test_video": {
    "logo_url": "https://cdn.jsdelivr.net/gh/youshandefeiyang/IPTV/logo/tg.jpg",
    "time_video_url": "https://cdn.jsdelivr.net/gh/youshandefeiyang/testvideo/time/time.mp4",
    "test_ad_url_1": "http://159.75.85.63:5680/d/ad/h264/playad.m3u8",
    "test_ad_url_2": "http://159.75.85.63:5680/d/ad/playad.m3u8"
  }
}
```

你可以根据需要修改此文件中的配置项。

## 一、推荐使用Docker一键运行，并配置watchtower监听Docker镜像更新，直接一劳永逸：

### 1，使用Docker一键配置allinone

```bash
# 拉取并运行容器
docker run -d --restart unless-stopped --privileged=true -p 35455:35455 --name allinone youshandefeiyang/allinone

# 或者，如果你有自己的 config.json 文件，可以挂载它
# docker run -d --restart unless-stopped --privileged=true -p 35455:35455 -v $(pwd)/config.json:/app/config.json --name allinone youshandefeiyang/allinone
```

### 2，一键配置watchtower每天凌晨两点自动监听allinone镜像更新，同步GitHub仓库：

```bash
docker run -d --name watchtower --restart unless-stopped -v /var/run/docker.sock:/var/run/docker.sock  containrrr/watchtower -c  --schedule "0 0 2 * * *"
```

### 3，使用 Docker Compose (可选)

你也可以使用 `docker-compose.yml` 文件来管理容器：

```yaml
version: '3'
services:
  iptv:
    image: registry.cn-hangzhou.aliyuncs.com/sky-devops/iptv:latest
    container_name: iptv
    privileged: true
    ports:
      - "35455:35455"
    volumes:
      - ./config.json:/app/config.json # 可选，挂载自定义配置文件
    restart: unless-stopped

  watchtower:
    image: containrrr/watchtower
    container_name: watchtower
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    command: --cleanup --schedule "0 0 2 * * *"
    restart: unless-stopped
```

然后运行以下命令启动服务：

```bash
docker-compose up -d
```

## 二、直接运行：

首先去action中下载对应平台二进制执行文件，然后解压并直接执行

```bash
chmod 777 allinone && ./allinone
```

建议搭配进程守护工具进行使用，windows直接双击运行！

## 三、详细使用方法

## **虎牙、斗鱼、YY实时M3U获取：**

### 虎牙一起看：

```
http://你的IP:35455/huyayqk.m3u
```

### 斗鱼一起看：

```
http://你的IP:35455/douyuyqk.m3u
```

### YY轮播：

```
http://你的IP:35455/yylunbo.m3u
```

### 如果使需要自定义M3U文件中的前缀域名，可以传入url参数（需要注意的是，当域名中含有特殊字符时，需要对链接进行urlencode处理）：

```
http://你的IP:35455/xxxyqk.m3u?url=http://192.168.10.1:35455
```

## **抖音：**

### 默认最高画质，浏览器打开并复制`(live.douyin.com/)xxxxxx`，只需要复制后面的xxxxx即可（可选flv和hls两种种流媒体传输方式，默认flv）：

```
http://你的IP:35455/douyin/xxxxx(?stream=hls)
```

## **斗鱼：**

### 1，可选m3u8和flv以及xs三种流媒体传输方式【`(www.douyu.com/)xxxxxx 或 (www.douyu.com/xx/xx?rid=)xxxxxx`，默认flv】：

```
http://你的IP:35455/douyu/xxxxx(?stream=flv)
```

## **BiliBili`(live.bilibili.com/)xxxxxx`：**

### 1，平台platform参数选择（默认web，如果有问题，可以切换h5平台）：

```
"web"   => "桌面端"
"h5"    => "h5端"
```

### 2，线路line参数选择（默认线路二，如果卡顿/看不了，请切换线路一或者三，一般直播间只会提供两条线路，所以建议线路一/二之间切换）：

```
"first"  => "线路一"
"second" => "线路二"
"third"  => "线路三"
```

### 3，画质quality参数选择（默认原画，可以看什么画质去直播间看看，能选什么画质就能加什么参数，参数错误一定不能播放）：

```
"30000" => "杜比"
"20000" => "4K"
"10000" => "原画"
"400"   => "蓝光"
"250"   => "超清"
"150"   => "高清"
"80"    => "流畅"
```

### 4，最后的代理链接示例：

```
http://你的IP:35455/bilibili/xxxxxx(?platform=h5&line=first&quality=10000)
```

## **虎牙`(huya.com/)xxxxxx`：**  

### 1，查看可用CDN：

```
http://你的IP:35455/huya/xxxxx?type=display
```

### 2，切换媒体类型（默认flv，可选flv、hls）： 

```
http://你的IP:35455/huya/xxxxx?media=hls
```

### 3，切换CDN（默认hwcdn，可选hycdn、alicdn、txcdn、hwcdn、hscdn、wscdn，具体可先访问1获取）：

```
http://你的IP:35455/huya/xxxxx?cdn=alicdn
```

### 4，最后的代理链接示例：

```
http://你的IP:35455/huya/xxxxx(?media=xxx&cdn=xxx)
```

## **YouTube:**

```
https://www.youtube.com/watch?v=cK4LemjoFd0
Rid: cK4LemjoFd0
http://你的IP:35455/youtube/cK4LemjoFd0(?quality=1080/720...)
```

## **YY（默认最高画质，参数为4）:**

```
https://www.yy.com/xxxx
http://你的IP:35455/yy/xxxx(?quality=1/2/3/4...)
```

## 更多平台后续会酌情添加

## **IPTV 代理功能 (新增)**

为了解决部分 IPTV 源仅限特定运营商网络内访问的问题，本项目新增了代理功能。该功能会将原始的 IPTV 源通过本服务器进行转发，使得所有用户都能访问。

### 获取代理版 IPTV M3U 列表

使用以下地址获取经过代理的 IPTV M3U 列表。列表中的每个频道链接都会指向本服务器的代理接口，再由代理接口转发到原始的运营商链接。

```
http://你的IP:35455/proxy.m3u
```

### 配置

代理功能默认是启用的。你可以在 `config.json` 文件中通过设置 `proxy_enabled` 为 `false` 来禁用它。

**重要**: 为了让播放器能够正确访问代理服务器，你必须在 `config.json` 中配置 `proxy_address` 项，填入你的服务器对外的完整访问地址（包含`http://`或`https://`以及端口号）。

```json
{
  // ... 其他配置项
  "proxy_enabled": true,
  "proxy_address": "http://你的公网IP或域名:35455"
}
```

**注意**: 启用代理功能会增加服务器的带宽消耗，因为所有的视频流都需要经过你的服务器进行转发。

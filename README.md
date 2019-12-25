### amber-login

[![Build Status](https://travis-ci.org/CharLemAznable/amber-login.svg?branch=master)](https://travis-ci.org/CharLemAznable/amber-login)
[![codecov](https://codecov.io/gh/CharLemAznable/amber-login/branch/master/graph/badge.svg)](https://codecov.io/gh/CharLemAznable/amber-login)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/CharLemAznable/amber-login)
[![MIT Licence](https://badges.frapsoft.com/os/mit/mit.svg?v=103)](https://opensource.org/licenses/mit-license.php)
![GitHub code size](https://img.shields.io/github/languages/code-size/CharLemAznable/amber-login)

统一登录.

#### 配置文件

1. ```appConfig.toml```

```toml
Port = 11325
ContextPath = ""
CookieDomain = "r.cn"   # 写入cookie的domain
CookieExpiredHours = 6  # cookie有效期(小时)
DevMode = false
```

2. ```logback.xml```

```xml
<logging>
    <filter enabled="true">
        <tag>file</tag>
        <type>file</type>
        <level>INFO</level>
        <property name="filename">sonar-qyhook.log</property>
        <property name="format">[%D %T] [%L] (%S) %M</property>
        <property name="rotate">false</property>
        <property name="maxsize">0M</property>
        <property name="maxlines">0K</property>
        <property name="daily">false</property>
    </filter>
</logging>
```

#### 部署执行

1. 下载最新的可执行文件压缩包并解压

    下载地址: [amber-login release](https://github.com/CharLemAznable/amber-login/releases)

```bash
$ tar -xvJf amber-login-[version].[arch].[os].tar.xz
```

2. 新建/编辑配置文件, 启动运行

```bash
$ nohup ./amber-login-[version].[arch].[os].bin &
```

#### 后台入口

  `/admin/index`

默认管理员:

  `admin/Admin18&$` `manage/Manage12#$`

#### 统一登录入口

  `/?appId={appId}&redirectUrl={redirectUrl}`

#### Golang Kits

  [amber](https://github.com/CharLemAznable/amber)

#### Java Kits

  [amber-java-kits](https://github.com/CharLemAznable/amber-java-kits).

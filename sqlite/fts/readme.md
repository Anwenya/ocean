## sqlite fts5 中文+拼音

使用该插件 https://github.com/wangfenjin/simple 实现中文及拼音搜索

go语言使用时需要加上编译参数开启fts5，不然go-sqlite3是不支持fts5的。

```bash

go run --tags fts5

```

插件的作者提供了两个函数解析关键词

```sql

-- 普通分词
simple_query("beicun") 
-- 结巴分词
jieba_query("北村")

```
# sqlite 加密

相关资料

1. 内嵌式sqlite3的cgo实现

   https://github.com/mattn/go-sqlite3

2. 提供对sqlite的加密（扩展）

   https://github.com/sqlcipher/sqlcipher

3. 提供了对sqlite的多种加密方式，兼容sqlcipher（扩展）

   https://github.com/utelle/SQLite3MultipleCiphers

4. 该仓库将 SQLite3 Multiple Ciphers 和 go-sqlite3 集成到了一个c文件中，方便使用。

   https://github.com/jgiannuzzi/go-sqlite3

   仓库作者在该issue中探讨了sqlite加密相关话题

   https://github.com/mattn/go-sqlite3/pull/1109

   仓库的本质应该是提供GO对SMC的访问吧。

## SQLite3MultipleCiphers

smc支持多种加密方式，也对sqlcipher兼容。

1. SQLCipher
    - 优点：
        - 强大的加密算法：使用 256-bit AES 加密，采用 CBC 模式和随机初始化向量（IV）。
        - 广泛支持：支持多种平台，包括 Windows、macOS、Linux、iOS 和 Android。
        - 社区支持：开源项目，有广泛的社区支持和文档。
    - 缺点：
        - 性能开销：由于使用强大的加密算法，可能会对性能有一定影响，尤其是在资源有限的设备上。
        - 文件大小：加密后的文件可能会比未加密的文件稍大。

2. ChaCha20
    - 优点：
        - 高性能：相比 AES，ChaCha20 在某些平台（特别是那些没有硬件加速的设备）上性能更好。
        - 安全性强：ChaCha20 被认为是非常安全的加密算法，广泛用于网络协议中。
    - 缺点：
        - 较少使用：相对于 AES，ChaCha20 的使用场景较少，因此支持和文档可能不如 AES 丰富。
        - 硬件加速支持有限：尽管性能优异，但缺乏广泛的硬件加速支持。

3. SEE (SQLite Encryption Extension)
    - 优点：
        - 官方支持：由 SQLite 官方提供和支持，兼容性和稳定性好。
        - 灵活配置：支持多种加密算法和模式，用户可以根据需求进行选择。
    - 缺点：
        - 商业许可：SEE 不是免费的，需要购买商业许可证才能使用。
        - 复杂性：配置和使用可能相对复杂，需要深入了解不同加密选项。

4. RC4 (不推荐)
    - 优点：
        - 简单快速：RC4 是一种流加密算法，实施起来非常简单且加密速度快。
        - 低资源消耗：适合资源有限的环境。
    - 缺点：
        - 安全性弱：RC4 被认为是不安全的，存在多种已知漏洞，不推荐用于敏感数据的加密。
        - 逐步淘汰：大多数现代加密库和标准都不再推荐或支持 RC4。

## 加密方式比较和选择

选择适合的加密方式需要考虑以下因素：

- 安全性：
    - 如果安全性是首要考虑，选择 SQLCipher (AES) 或 ChaCha20。
- 性能：
    - 在资源有限或没有硬件加速的设备上，ChaCha20 可能是更好的选择。
    - 如果有硬件加速支持，SQLCipher (AES) 是不错的选择。
- 兼容性和支持：
    - SQLCipher 有广泛的社区支持和文档，适合需要跨平台支持的项目。
    - SEE 由 SQLite 官方提供支持，适合商业项目。
- 成本：
    - 如果预算有限，优先考虑开源的 SQLCipher 和 ChaCha20。
    - SEE 需要购买许可证，适合有预算并需要官方支持的项目。

## 使用GO连接数据库

"github.com/jgiannuzzi/go-sqlite3"
在GO层面的使用和之前完全一致，只是要使用加密功能需要指定一些参数，本质上也是通过PRAGMA命令的方式来设置一些配置。官方不推荐通过URI的方式来加密，因为这些值在内存中是可见的。
但通过GO目前只能使用该方式。（除非自己改源码）

详细内容还是要看SMC的官方文档，https://utelle.github.io/SQLite3MultipleCiphers/docs/configuration/config_uri/

```go

// Encryption key
if key != "" {
   if err := exec(fmt.Sprintf("PRAGMA key = %s;", key)); err != nil {
      C.sqlite3_close_v2(db)
        return nil, err
   }
}

```

```go

import (
    _ "github.com/jgiannuzzi/go-sqlite3"
)

sql.Open("sqlite3", "test_cipher_1.db?_cipher=sqlcipher&_key=123456")

```

"github.com/jgiannuzzi/go-sqlite3" 是在 "github.com/mattn/go-sqlite3" 的基础上结合了 SMC 的加密功能。

同时导入这两个库是会有冲突的（两个库会注册相同的CGO函数，毕竟除了加密功能都和之前一致），如果使用sql.Open来操作数据库的话，不用担心这个问题（用哪个就只导入哪个）。

```go

_ "github.com/mattn/go-sqlite3" // 默认

_ "github.com/jgiannuzzi/go-sqlite3" // 支持加密

```

如果要用GORM的话，就有问题了。因为GORM在 "gorm.io/driver/sqlite" 中会导入 "github.com/mattn/go-sqlite3"
来注册驱动。而我们再导入 "github.com/jgiannuzzi/go-sqlite3" 的话就会有冲突了。
要解决这个问题可以在mod中替换掉 github.com/mattn/go-sqlite3，这样实际就使用
github.com/jgiannuzzi/go-sqlite3了，使用的就是支持加密的版本了。

```go

require gorm.io/gorm v1.25.10

replace github.com/mattn/go -sqlite3 => github.com/jgiannuzzi/go -sqlite3 v1.14.17-0.20240122133042-fb824c8e339e

```

创建加密数据库测试

```go

// 不指定版本号默认使用AES-256bit加密
dsn := "file:test_cipher_1.db?_cipher=sqlcipher&_key=123456"

// 指定版本号使用对应版本的加密算法 版本4是SHA512
dsn := "file:test_cipher_1.db?_cipher=sqlcipher&_key=123456&_legacy=4"

// 使用chacha20算法
dsn := "file:test_cipher_chacha20.db?_cipher=chacha20&_key=123456"

db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})

```

要用带界面的客户端打开加密的数据库，推荐使用SQLiteStudio，DB Browser支持不好。



## 产品环境编译

codecgen工具利用go generate方式实现代码生成，减少因为默认模式下因为go反射对性能的影响。

go get -u github.com/ugorji/go/codec/codecgen
go generate vcs.taiyouxi.net/gamex/logics

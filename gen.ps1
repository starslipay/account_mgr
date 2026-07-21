# 生成rpc框架代码
goctl rpc protoc account_mgr.proto --go_out=. --go-grpc_out=. --zrpc_out=.

pushd
cd model/mysql
# 生成mysql模型代码
goctl model mysql ddl -src account.sql -dir .
popd

$VERSION = "v1.0.0"

pushd ..

# 先删容器，避免名称冲突
docker rm -f account_mgr
docker rmi -f account_mgr:$VERSION
docker build -t account_mgr:$VERSION .
docker run -d --name account_mgr --network local_deps_install_dev_net -p 30881:8080 account_mgr:$VERSION
docker ps
docker logs account_mgr

popd
### 安装 kubebuilder
```
git clone https://github.com/kubernetes-sigs/kubebuilder
cd kubebuilder
make build
cp kubebuilder $GOPATH/bin/
```
### 安装 kustomize
```
git clone https://github.com/kubernetes-sigs/kustomize.git
cd ./kustomize/kustomize
go build main.go
cp main $GOPATH/bin/kustomize
```
### 生成代码
```
mkdir $GOPATH/kubenebula.io/kubenebula -p
cd $GOPATH/kubenebula.io/kubenebula
go mod init kubenebula.io/kubenebula
kubebuilder init --domain kubenebula.io --license apache2 --owner "The KubeNebula authors"
kubebuilder create api --group tenant --version v1alpha1 --kind Team
```
### 修改crd参数
修改 `api/v1alpha1/team_types.go`
然后 `make`
### 修改controller
修改 `controllers/team_controller.go`的`Reconcile`函数，添加逻辑代码